package run

import (
	"context"
	"fmt"
	"github.com/google/go-github/v47/github"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/giturl"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spring-financial-group/mqa-helpers/pkg/cobras/helper"
	"github.com/spring-financial-group/mqa-helpers/pkg/cobras/templates"

	"github.com/spring-financial-group/peacock/pkg/config"
	"github.com/spring-financial-group/peacock/pkg/domain"
	"github.com/spring-financial-group/peacock/pkg/git"
	"github.com/spring-financial-group/peacock/pkg/handlers"
	"github.com/spring-financial-group/peacock/pkg/handlers/email"
	"github.com/spring-financial-group/peacock/pkg/handlers/slack"
	"github.com/spring-financial-group/peacock/pkg/message"
	"github.com/spring-financial-group/peacock/pkg/rootcmd"
	"github.com/spring-financial-group/peacock/pkg/utils"
	"strings"
)

// Options for the run command
type Options struct {
	PRNumber     int
	GitServerURL string
	GitToken     string
	Owner        string
	RepoName     string

	DryRun bool

	SlackToken string

	SmtpHost     string
	SmtpUsername string
	SmtpPassword string
	SmtpPort     int

	pr *github.PullRequest

	Git      domain.Git
	Handlers map[string]domain.MessageHandler
	Config   *config.Config
}

var (
	longDesc = templates.LongDesc(`
		run notifies teams of new release information. The messages are taken for the body of a pull request and the recipient
		info is taken from a config file in the repository.
`)

	example = templates.Examples(`
		%s run [flags]
	`)
)

func NewCmdRun() *cobra.Command {
	o := &Options{}
	cmd := &cobra.Command{
		Use:     "run",
		Short:   "Notifies teams of new release information",
		Long:    longDesc,
		Example: fmt.Sprintf(example, rootcmd.BinaryName),
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			err := o.Run()
			helper.CheckErr(err)
		},
	}
	// Git flags
	cmd.Flags().StringVarP(&o.GitServerURL, "git-server", "", "", fmt.Sprintf("the git server URL to create the scm client. Default is %s", giturl.GitHubURL))
	cmd.Flags().IntVarP(&o.PRNumber, "pr-number", "p", -1, "github auth token")
	cmd.Flags().StringVarP(&o.GitToken, "git-token", "", "", "the git token used to operate on the git repository. If not specified it's loaded from the git credentials file")
	cmd.Flags().StringVarP(&o.Owner, "git-owner", "", "", "the owner of the git repository. If not specified it's loaded from the local git repo")
	cmd.Flags().StringVarP(&o.RepoName, "git-repo", "", "", "the name of the git repo to run on. If not specified it's loaded from the local git repo")

	// Command specific flags
	cmd.Flags().BoolVarP(&o.DryRun, "dry-run", "", false, "parses the messages and config, returning validation as a comment on the pr. Does not send messages. PR number is required for this. Default is false")

	// Slack flags
	cmd.Flags().StringVarP(&o.SlackToken, "slack-token", "", "", "the slack token used to send the messages to slack channels")

	// Email flags
	cmd.Flags().StringVarP(&o.SmtpHost, "smtp-host", "", "", "the host SMTP server")
	cmd.Flags().StringVarP(&o.SmtpUsername, "smtp-username", "", "", "the username to connect to the SMTP server")
	cmd.Flags().StringVarP(&o.SmtpPassword, "smtp-password", "", "", "the password to connect to the SMTP server")
	cmd.Flags().IntVarP(&o.SmtpPort, "smtp-port", "", -1, "the port of the SMTP server")

	return cmd
}

func (o *Options) Run() error {
	log.Logger().Info("Initialising variables & clients")
	err := o.initialiseFlagsAndClients()
	if err != nil {
		return errors.Wrap(err, "failed to validate input args & clients")
	}

	ctx := context.Background()
	err = o.GetPullRequest(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get pull request")
	}

	if o.Config == nil {
		log.Logger().Info("Loading config from local instance")
		o.Config, err = config.LoadConfig()
		if err != nil {
			err = errors.Wrapf(err, "failed to load config")
			o.PostErrorToPR(ctx, err)
			return err
		}
	}

	if o.Handlers == nil {
		log.Logger().Info("Initialising message handlers")
		err = o.initialiseHandlers()
		if err != nil {
			err = errors.Wrapf(err, "failed to init handlers")
			o.PostErrorToPR(ctx, err)
			return err
		}
	}

	log.Logger().Info("Parsing messages from pull request body")
	messages, err := message.ParseMessagesFromMarkdown(*o.pr.Body)
	if err != nil {
		err = errors.Wrapf(err, "failed to parse messages from pull request")
		o.PostErrorToPR(ctx, err)
		return err
	}

	log.Logger().Info("Validating messages")
	err = o.ValidateMessagesWithConfig(messages)
	if err != nil {
		err = errors.Wrapf(err, "failed validate messages with config")
		o.PostErrorToPR(ctx, err)
		return err
	}

	if o.DryRun {
		log.Logger().Info("Posting message breakdown to pull request")
		breakdown, err := o.GenerateMessageBreakdown(messages)
		if err != nil {
			err = errors.Wrapf(err, "failed to generate breakdown of messages")
			o.PostErrorToPR(ctx, err)
			return err
		}
		log.Logger().Info(breakdown)
		// Return before sending messages
		return o.Git.CommentOnPR(ctx, o.pr, breakdown)
	}

	log.Logger().Info("Sending messages")
	err = o.SendMessages(messages)
	if err != nil {
		return err
	}
	return nil
}

func (o *Options) GetPullRequest(ctx context.Context) (err error) {
	if o.DryRun {
		// If it's a dry run we need to be given the pr number that we're in
		log.Logger().Info("Getting pull request from PR number")
		o.pr, err = o.Git.GetPullRequestFromPRNumber(ctx, o.PRNumber)
	} else {
		// If not then we can get it from the last commit in the local instance
		log.Logger().Info("Getting pull request from last commit")
		o.pr, err = o.Git.GetPullRequestFromLastCommit(ctx)
	}
	if err != nil {
		return errors.Wrap(err, "failed to get pull request")
	}

	// We should check whether this pr is usable
	if o.pr.Body == nil {
		return errors.New("no body found in pull request")
	}
	return nil
}

// SendMessages send the messages using the message handlers
func (o *Options) SendMessages(messages []message.Message) error {
	var errs []error
	for _, m := range messages {
		err := o.sendMessage(m)
		if err != nil {
			log.Logger().Error(err)
			errs = append(errs, err)
			continue
		}
	}
	if len(errs) > 0 {
		return errors.New("failed to send messages")
	}
	return nil
}

func (o *Options) sendMessage(message message.Message) error {
	team := o.Config.GetTeamByName(message.TeamName)
	err := o.Handlers[team.ContactType].Send(message.Content, team.Addresses)
	if err != nil {
		return errors.Wrapf(err, "failed to send messages to %s using %s", team.Name, team.ContactType)
	}
	log.Logger().Infof("Message successfully sent to %s via %s\n", team.Name, team.Addresses)
	return nil
}

// ValidateMessagesWithConfig checks that the messages found in the pr meet the requirements of the config
func (o *Options) ValidateMessagesWithConfig(messages []message.Message) error {
	allTeamsInConfig := o.Config.GetAllTeamNames()
	for _, m := range messages {
		// Check the team name actually exists in config
		exist := utils.ExistsInSlice(m.TeamName, allTeamsInConfig)
		if !exist {
			return errors.Errorf("Team %s does not exist in config.yaml:\n%v", m.TeamName, allTeamsInConfig)
		}

		// Check that the handler for the teams contact type is initialised
		team := o.Config.GetTeamByName(m.TeamName)
		if o.Handlers[team.ContactType] == nil {
			return errors.Errorf("Team \"%s\" has contact type \"%s\", handler not initialised for this type - check input flags", m.TeamName, team.ContactType)
		}
	}
	return nil
}

// GenerateMessageBreakdown creates a breakdown of the messages found in the pr description
func (o *Options) GenerateMessageBreakdown(messages []message.Message) (string, error) {
	allTeamsInConfig := o.Config.GetAllTeamNames()
	breakDown := fmt.Sprintf(
		"### Validation\nSuccessfully parsed %d message(s)\n%d/%d teams in config to notify\n",
		len(messages), len(messages), len(allTeamsInConfig),
	)

	for i, m := range messages {
		team := o.Config.GetTeamByName(m.TeamName)
		newMessage := fmt.Sprintf("***\n"+
			"### Message [%d/%d]\n#### Team: %s\n#### Contact Type: %s\n#### Content:\n%s\n",
			i+1, len(messages), m.TeamName, team.ContactType, m.Content,
		)
		breakDown = breakDown + newMessage
	}
	return strings.TrimSpace(breakDown), nil
}

// PostErrorToPR posts an error to the pull request as a comment
func (o *Options) PostErrorToPR(ctx context.Context, err error) {
	// If it's not a DryRun then we shouldn't post the error back to the pr
	if o.DryRun {
		err = o.Git.CommentOnPR(ctx, o.pr, "Error: "+err.Error())
		if err != nil {
			panic(err)
		}
	}
}

// initialiseFlagsAndClients checks that all the variables required to run the command are set up correctly
// and sets up the required clients
func (o *Options) initialiseFlagsAndClients() (err error) {
	// validate flags
	if o.DryRun && o.PRNumber == -1 {
		return errors.New("pr-number required")
	}
	if o.GitServerURL == "" {
		o.GitServerURL = giturl.GitHubURL
	}

	// Init git clients
	if o.Git == nil {
		o.Git, err = git.NewClient(o.GitServerURL, o.GitToken, o.Owner, o.RepoName)
		if err != nil {
			return errors.Wrap(err, "failed to initialise git clients")
		}
	}
	return nil
}

// initialiseHandlers initialises the message handlers depending on the flags passed through to the command.
// It then checks that all the handlers required by the config have been initialised.
func (o *Options) initialiseHandlers() (err error) {
	o.Handlers = map[string]domain.MessageHandler{}
	if o.SlackToken != "" {
		o.Handlers[handlers.Slack], err = slack.NewSlackHandler(o.SlackToken)
		if err != nil {
			return errors.Wrap(err, "failed to initialise Slack handler")
		}
	}
	if o.SmtpHost != "" && o.SmtpUsername != "" && o.SmtpPassword != "" && o.SmtpPort != -1 {
		o.Handlers[handlers.Email], err = email.NewEmailHandler(o.SmtpPort, o.SmtpHost, o.SmtpUsername, o.SmtpPassword)
		if err != nil {
			return errors.Wrap(err, "failed to initialise Email handler")
		}
	}

	// We should check that all the handlers required by the config have been initialised
	for _, t := range o.Config.GetAllContactTypes() {
		if o.Handlers[t] == nil {
			return errors.Errorf(
				"contact type \"%s\" found in config but no handler has been initialised, "+
					"check required flags have been passed for this type", t)
		}
	}
	return nil
}
