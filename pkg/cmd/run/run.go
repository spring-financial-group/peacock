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
	"github.com/spring-financial-group/peacock/pkg/domain"
	"github.com/spring-financial-group/peacock/pkg/feathers"
	"github.com/spring-financial-group/peacock/pkg/git"
	"github.com/spring-financial-group/peacock/pkg/handlers"
	"github.com/spring-financial-group/peacock/pkg/handlers/slack"
	"github.com/spring-financial-group/peacock/pkg/message"
	"github.com/spring-financial-group/peacock/pkg/rootcmd"
	"github.com/spring-financial-group/peacock/pkg/utils"
	"os"
	"strconv"
	"strings"
)

// Options for the run command
type Options struct {
	PRNumber     int
	GitServerURL string
	GitHubToken  string
	RepoOwner    string
	RepoName     string

	SlackToken string

	DryRun            bool
	CommentValidation bool

	pr *github.PullRequest

	Git      domain.Git
	Handlers map[string]domain.MessageHandler
	Config   *feathers.Feathers
}

var (
	longDesc = templates.LongDesc(`
		run notifies teams of new release information. The messages are taken for the body of a pull request and the recipient
		info is taken from the feathers file in the repository.
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

	err := o.ParseEnvVars(cmd)
	helper.CheckErr(err)

	// Command specific flags
	cmd.Flags().BoolVarP(&o.DryRun, "dry-run", "", false, "parses the messages and feathers, returning validation as a comment on the pr. Does not send messages. PR number is required for this. Default is false")
	cmd.Flags().BoolVarP(&o.CommentValidation, "comment-validation", "", false, "posts a comment to the pr with the validation results if successful. Default is false.")

	return cmd
}

// ParseEnvVars uses the flags passed to the command to overwrite the default environment variable keys. Then loads the
// environment variables.
func (o *Options) ParseEnvVars(cmd *cobra.Command) (err error) {
	keys := struct {
		PRNumberKey     string
		GitServerURLKey string
		GitHubKey       string
		RepoOwnerKey    string
		RepoNameKey     string
		SlackTokenKey   string
	}{}

	// Flags to overwrite default environment variable keys
	cmd.Flags().StringVarP(&keys.GitServerURLKey, "git-server-key", "", "GIT_SERVER", fmt.Sprintf("the environment variable key for the git server URL. If no env var is passed then default is %s", giturl.GitHubURL))
	cmd.Flags().StringVarP(&keys.PRNumberKey, "pr-number-key", "p", "PULL_NUMBER", "the environment variable key for the pull request number that peacock is running on.")
	cmd.Flags().StringVarP(&keys.GitHubKey, "git-token-key", "", "GITHUB_TOKEN", "the environment variable key for the git token used to operate on the git repository.")
	cmd.Flags().StringVarP(&keys.RepoOwnerKey, "git-owner-key", "", "REPO_OWNER", "the environment variable key for the owner of the git repository.")
	cmd.Flags().StringVarP(&keys.RepoNameKey, "git-repo-key", "", "REPO_NAME", "the environment variable key for the name of the git repo to run on.")
	cmd.Flags().StringVarP(&keys.SlackTokenKey, "slack-token-key", "", "SLACK_TOKEN", "the environment variable key for the slack token used to send the messages to slack channels")

	o.PRNumber = -1
	if prNumber := os.Getenv(keys.PRNumberKey); prNumber != "" {
		o.PRNumber, err = strconv.Atoi(prNumber)
		if err != nil {
			return err
		}
	}

	o.GitServerURL = os.Getenv(keys.GitServerURLKey)
	o.GitHubToken = os.Getenv(keys.GitHubKey)
	o.RepoOwner = os.Getenv(keys.RepoOwnerKey)
	o.RepoName = os.Getenv(keys.RepoNameKey)
	o.SlackToken = os.Getenv(keys.SlackTokenKey)
	return nil
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
		log.Logger().Info("Loading feathers from local instance")
		o.Config, err = feathers.LoadConfig()
		if err != nil {
			err = errors.Wrapf(err, "failed to load feathers")
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

	// If no messages then we should exit with 0 code
	if messages == nil {
		return nil
	}

	log.Logger().Info("Validating messages")
	err = o.ValidateMessagesWithConfig(messages)
	if err != nil {
		err = errors.Wrapf(err, "failed validate messages with feathers")
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
		if o.CommentValidation {
			if err := o.Git.CommentOnPR(ctx, o.pr, breakdown); err != nil {
				return err
			}
		}
		return nil
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
	teams := o.Config.GetTeamsByNames(message.TeamNames...)
	for _, team := range teams {
		err := o.Handlers[team.ContactType].Send(message.Content, team.Addresses)
		if err != nil {
			return errors.Wrapf(err, "failed to send messages to %s using %s", team.Name, team.ContactType)
		}
		log.Logger().Infof("Message successfully sent to %s via %s\n", team.Name, team.Addresses)
	}
	return nil
}

// ValidateMessagesWithConfig checks that the messages found in the pr meet the requirements of the feathers
func (o *Options) ValidateMessagesWithConfig(messages []message.Message) error {
	allTeamsInConfig := o.Config.GetAllTeamNames()
	for _, m := range messages {
		// Check the team name actually exists in feathers
		for _, msgTeamName := range m.TeamNames {
			exist := utils.ExistsInSlice(msgTeamName, allTeamsInConfig)
			if !exist {
				return errors.Errorf("Team %s does not exist in feathers.yaml:\n%v", msgTeamName, allTeamsInConfig)
			}
		}

		// Check that the handler for the teams contact type is initialised
		teams := o.Config.GetTeamsByNames(m.TeamNames...)
		for _, team := range teams {
			if o.Handlers[team.ContactType] == nil {
				return errors.Errorf("Team \"%s\" has contact type \"%s\", handler not initialised for this type - check input flags", team.Name, team.ContactType)
			}
		}
	}
	return nil
}

// GenerateMessageBreakdown creates a breakdown of the messages found in the pr description
func (o *Options) GenerateMessageBreakdown(messages []message.Message) (string, error) {
	allTeamsInConfig := o.Config.GetAllTeamNames()
	breakDown := fmt.Sprintf(
		"### Validation\nSuccessfully parsed %d message(s)\n%d/%d teams in feathers to notify\n",
		len(messages), len(messages), len(allTeamsInConfig),
	)

	for i, m := range messages {
		contactTypes := o.Config.GetContactTypesByTeamNames(m.TeamNames...)
		newMessage := fmt.Sprintf("***\n"+
			"### Message [%d/%d]\n#### Teams: %s\n#### Contact Types: %s\n#### Content:\n%s\n",
			i+1, len(messages), m.TeamNames, contactTypes, m.Content,
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
		o.Git, err = git.NewClient(o.GitServerURL, o.RepoOwner, o.RepoName, o.GitHubToken)
		if err != nil {
			return errors.Wrap(err, "failed to initialise git clients")
		}
	}
	return nil
}

// initialiseHandlers initialises the message handlers depending on the flags passed through to the command.
// It then checks that all the handlers required by the feathers have been initialised.
func (o *Options) initialiseHandlers() (err error) {
	o.Handlers = map[string]domain.MessageHandler{}
	if o.SlackToken != "" {
		o.Handlers[handlers.Slack], err = slack.NewSlackHandler(o.SlackToken)
		if err != nil {
			return errors.Wrap(err, "failed to initialise Slack handler")
		}
	}

	// We should check that all the handlers required by the feathers have been initialised
	for _, t := range o.Config.GetAllContactTypes() {
		if o.Handlers[t] == nil {
			return errors.Errorf(
				"contact type \"%s\" found in feathers but no handler has been initialised, "+
					"check required flags have been passed for this type", t)
		}
	}
	return nil
}
