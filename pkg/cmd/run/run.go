package run

import (
	"context"
	"fmt"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/giturl"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spring-financial-group/mqa-helpers/pkg/cobras/helper"
	"github.com/spring-financial-group/mqa-helpers/pkg/cobras/templates"
	"github.com/spring-financial-group/peacock/pkg/domain"
	"github.com/spring-financial-group/peacock/pkg/feathers"
	"github.com/spring-financial-group/peacock/pkg/git"
	"github.com/spring-financial-group/peacock/pkg/git/github"
	"github.com/spring-financial-group/peacock/pkg/handlers"
	"github.com/spring-financial-group/peacock/pkg/handlers/slack"
	"github.com/spring-financial-group/peacock/pkg/handlers/webhook"
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

	WebhookURL    string
	WebhookToken  string
	WebhookSecret string

	DryRun            bool
	CommentValidation bool
	Subject           string

	GitServerClient domain.GitServer
	Git             domain.Git
	Handlers        map[string]domain.MessageHandler
	Config          *feathers.Feathers
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
	cmd.Flags().StringVarP(&o.Subject, "subject", "", "", "a subject to add to the messages for the handlers that require it. If empty then a subject will be generated.")

	return cmd
}

// ParseEnvVars uses the flags passed to the command to overwrite the default environment variable keys. Then loads the
// environment variables.
func (o *Options) ParseEnvVars(cmd *cobra.Command) (err error) {
	keys := struct {
		PRNumber         string
		GitServerURL     string
		GitHubToken      string
		RepoOwner        string
		RepoName         string
		SlackToken       string
		WebhookURL       string
		WebhookAuthToken string
		WebhookSecret    string
	}{}

	// Flags to overwrite default environment variable keys
	cmd.Flags().StringVarP(&keys.GitServerURL, "git-server-key", "", "GIT_SERVER", fmt.Sprintf("the environment variable key for the git server URL. If no env var is passed then default is %s", giturl.GitHubURL))
	cmd.Flags().StringVarP(&keys.PRNumber, "pr-number-key", "p", "PULL_NUMBER", "the environment variable key for the pull request number that peacock is running on.")
	cmd.Flags().StringVarP(&keys.GitHubToken, "git-token-key", "", "GITHUB_TOKEN", "the environment variable key for the git token used to operate on the git repository.")
	cmd.Flags().StringVarP(&keys.RepoOwner, "git-owner-key", "", "REPO_OWNER", "the environment variable key for the owner of the git repository.")
	cmd.Flags().StringVarP(&keys.RepoName, "git-repo-key", "", "REPO_NAME", "the environment variable key for the name of the git repo to run on.")
	cmd.Flags().StringVarP(&keys.SlackToken, "slack-token-key", "", "SLACK_TOKEN", "the environment variable key for the slack token used to send the messages to slack channels")
	cmd.Flags().StringVarP(&keys.WebhookURL, "webhook-URL-key", "", "WEBHOOK_URL", "the environment variable key for the webhook URL")
	cmd.Flags().StringVarP(&keys.WebhookAuthToken, "webhook-auth-token-key", "", "WEBHOOK_AUTH_TOKEN", "the environment variable key for the webhook auth token")
	cmd.Flags().StringVarP(&keys.WebhookSecret, "webhook-HMAC-secret-key", "", "WEBHOOK_SECRET", "the environment variable key for the webhook HMAC secret")

	o.PRNumber = -1
	if prNumber := os.Getenv(keys.PRNumber); prNumber != "" {
		o.PRNumber, err = strconv.Atoi(prNumber)
		if err != nil {
			return err
		}
	}

	o.GitServerURL = os.Getenv(keys.GitServerURL)
	o.GitHubToken = os.Getenv(keys.GitHubToken)
	o.RepoOwner = os.Getenv(keys.RepoOwner)
	o.RepoName = os.Getenv(keys.RepoName)
	o.SlackToken = os.Getenv(keys.SlackToken)
	o.WebhookURL = os.Getenv(keys.WebhookURL)
	o.WebhookToken = os.Getenv(keys.WebhookAuthToken)
	o.WebhookSecret = os.Getenv(keys.WebhookSecret)
	return nil
}

func (o *Options) Run() error {
	log.Logger().Info("Initialising variables & clients")
	err := o.initialiseFlagsAndClients()
	if err != nil {
		return errors.Wrap(err, "failed to validate input args & clients")
	}

	ctx := context.Background()
	prBody, err := o.GetPullRequestBody(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get pull request")
	}
	// We should check that the body actually exists
	if prBody == nil {
		log.Logger().Infof("No Body found for PR%d, exiting", o.PRNumber)
		return nil
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
	messages, err := message.ParseMessagesFromMarkdown(*prBody)
	if err != nil {
		err = errors.Wrapf(err, "failed to parse messages from pull request")
		o.PostErrorToPR(ctx, err)
		return err
	}
	// If no messages then we should exit with 0 code
	if messages == nil {
		log.Logger().Info("No messages found in markdown, exiting")
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
			if err := o.GitServerClient.CommentOnPR(ctx, o.PRNumber, breakdown); err != nil {
				return err
			}
		}
		return nil
	}

	// Some message handlers use subjects, if one isn't passed then we should generate it
	if o.Subject == "" {
		o.GenerateSubject()
	}

	log.Logger().Info("Sending messages")
	err = o.SendMessages(messages)
	if err != nil {
		return err
	}
	return nil
}

func (o *Options) GetPullRequestBody(ctx context.Context) (*string, error) {
	var err error
	var body *string
	if o.DryRun {
		// If it's a dry run we need to be given the pr number that we're in
		log.Logger().Info("Getting pull request from PR number")
		body, err = o.GitServerClient.GetPullRequestBodyFromPRNumber(ctx, o.PRNumber)
	} else {
		// If not then we can get it from the last commit in the local instance
		log.Logger().Info("Getting pull request from last commit")
		sha, err := o.Git.GetLatestCommitSHA()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get latest commit sha")
		}
		body, err = o.GitServerClient.GetPullRequestBodyFromCommit(ctx, sha)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get pull request")
	}
	return body, nil
}

func (o *Options) GenerateSubject() {
	o.Subject = fmt.Sprintf("New Release Notes for %s", o.RepoName)
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
		err := o.Handlers[team.ContactType].Send(message.Content, o.Subject, team.Addresses)
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
		err = o.GitServerClient.CommentOnPR(ctx, o.PRNumber, "Error: "+err.Error())
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
		o.Git = git.NewClient()
	}
	if o.GitServerClient == nil {
		o.GitServerClient = github.NewClient(o.RepoOwner, o.RepoName, o.GitHubToken)
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

	if o.WebhookURL != "" {
		o.Handlers[handlers.Webhook] = webhook.NewWebHookHandler(o.WebhookURL, o.WebhookToken, o.WebhookSecret)
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
