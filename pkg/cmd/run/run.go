package run

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spring-financial-group/peacock/pkg/config"
	"github.com/spring-financial-group/peacock/pkg/domain"
	"github.com/spring-financial-group/peacock/pkg/feathers"
	"github.com/spring-financial-group/peacock/pkg/git"
	"github.com/spring-financial-group/peacock/pkg/git/comment"
	"github.com/spring-financial-group/peacock/pkg/git/github"
	"github.com/spring-financial-group/peacock/pkg/message"
	"github.com/spring-financial-group/peacock/pkg/msgclient"
	"github.com/spring-financial-group/peacock/pkg/rootcmd"
	"github.com/spring-financial-group/peacock/pkg/utils"
	"github.com/spring-financial-group/peacock/pkg/utils/templates"
	"os"
	"strconv"
)

// Options for the run command
type Options struct {
	PRNumber     int
	GitServerURL string
	GitHubToken  string
	GitUser      string
	RepoOwner    string
	RepoName     string

	SlackToken string

	WebhookURL    string
	WebhookToken  string
	WebhookSecret string

	DryRun            bool
	CommentValidation bool
	Subject           string

	GitServerClient domain.SCM
	Git             domain.Git
	MSGHandler      domain.MessageHandler
	Feathers        *feathers.Feathers
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
			utils.CheckErr(err)
		},
	}

	err := o.ParseEnvVars(cmd)
	utils.CheckErr(err)

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
		GitUser          string
		SlackToken       string
		WebhookURL       string
		WebhookAuthToken string
		WebhookSecret    string
	}{}

	// Flags to overwrite default environment variable keys
	cmd.Flags().StringVarP(&keys.GitServerURL, "git-server-key", "", "GIT_SERVER", fmt.Sprintf("the environment variable key for the git server URL. If no env var is passed then default is %s", domain.GitHubURL))
	cmd.Flags().StringVarP(&keys.PRNumber, "pr-number-key", "p", "PULL_NUMBER", "the environment variable key for the pull request number that peacock is running on.")
	cmd.Flags().StringVarP(&keys.GitHubToken, "git-token-key", "", "GITHUB_TOKEN", "the environment variable key for the git token used to operate on the git repository.")
	cmd.Flags().StringVarP(&keys.RepoOwner, "git-owner-key", "", "REPO_OWNER", "the environment variable key for the owner of the git repository.")
	cmd.Flags().StringVarP(&keys.RepoName, "git-repo-key", "", "REPO_NAME", "the environment variable key for the name of the git repo to run on.")
	cmd.Flags().StringVarP(&keys.RepoName, "git-user", "", "GIT_USER", "the environment variable key for the name user to interact with Git.")
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
	o.GitUser = os.Getenv(keys.GitUser)
	o.SlackToken = os.Getenv(keys.SlackToken)
	o.WebhookURL = os.Getenv(keys.WebhookURL)
	o.WebhookToken = os.Getenv(keys.WebhookAuthToken)
	o.WebhookSecret = os.Getenv(keys.WebhookSecret)
	return nil
}

func (o *Options) Run() error {
	log.Info("Initialising variables & clients")
	err := o.initialiseFlagsAndClients()
	if err != nil {
		return errors.Wrap(err, "failed to validate input args & clients")
	}

	ctx := context.Background()
	prBody, err := o.GetPullRequestBody(ctx)
	if err != nil {
		return err
	}
	// We should check that the body actually exists
	if prBody == nil {
		log.Infof("No Body found for PR%d, exiting", o.PRNumber)
		return nil
	}

	if o.Feathers == nil {
		log.Info("Loading feathers from local instance")
		o.Feathers, err = feathers.GetFeathersFromFile()
		if err != nil {
			err = errors.Wrapf(err, "failed to load feathers")
			o.PostErrorToPR(ctx, err)
			return err
		}
	}

	if o.MSGHandler == nil {
		o.MSGHandler = msgclient.NewMessageHandler(&config.MessageHandlers{
			Slack: config.Slack{
				Token: o.SlackToken,
			},
			Webhook: config.Webhook{
				URL:    o.WebhookURL,
				Token:  o.WebhookToken,
				Secret: o.WebhookSecret,
			},
		})
	}

	log.Info("Parsing messages from pull request body")
	messages, err := message.ParseMessagesFromMarkdown(*prBody)
	if err != nil {
		err = errors.Wrapf(err, "failed to parse messages from pull request")
		o.PostErrorToPR(ctx, err)
		return err
	}
	// If no messages then we should exit with 0 code
	if messages == nil {
		log.Info("No messages found in markdown, exiting")
		return nil
	}

	log.Info("Validating messages")
	err = o.ValidateMessagesWithConfig(messages)
	if err != nil {
		err = errors.Wrapf(err, "failed validate messages with feathers")
		o.PostErrorToPR(ctx, err)
		return err
	}

	if o.DryRun {
		log.Info("Generating message breakdown")
		breakdown, err := o.GetMessageBreakdown(ctx, messages)
		if err != nil {
			err = errors.Wrapf(err, "failed to generate breakdown of messages")
			o.PostErrorToPR(ctx, err)
			return err
		}
		// Return before sending messages
		if o.CommentValidation && breakdown != "" {
			if err := o.GitServerClient.CommentOnPR(ctx, breakdown); err != nil {
				return err
			}
		}
		return nil
	}

	// Some message handlers use subjects, if one isn't passed then we should generate it
	if o.Subject == "" {
		o.GenerateSubject()
	}

	log.Info("Sending messages")
	err = o.MSGHandler.SendMessages(o.Feathers, messages)
	if err != nil {
		return err
	}
	return nil
}

func (o *Options) GetPullRequestBody(ctx context.Context) (*string, error) {
	var err error
	var body *string
	var sha string
	if o.DryRun {
		// If it's a dry run we need to be given the pr number that we're in
		log.Info("Getting pull request from PR number")
		body, err = o.GitServerClient.GetPullRequestBodyFromPRNumber(ctx)
	} else {
		// If not then we can get it from the last commit in the local instance
		log.Info("Getting pull request from last commit")
		sha, err = o.Git.GetLatestCommitSHA("")
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

// ValidateMessagesWithConfig checks that the messages found in the pr meet the requirements of the feathers
func (o *Options) ValidateMessagesWithConfig(messages []message.Message) error {
	allTeamsInConfig := o.Feathers.GetAllTeamNames()
	for _, m := range messages {
		// Check the team name actually exists in feathers
		for _, msgTeamName := range m.TeamNames {
			exist := utils.ExistsInSlice(msgTeamName, allTeamsInConfig)
			if !exist {
				return errors.Errorf("Team %s does not exist in feathers.yaml:\n%v", msgTeamName, allTeamsInConfig)
			}
		}

		// Check that the handler for the teams contact type is initialised
		teams := o.Feathers.GetTeamsByNames(m.TeamNames...)
		for _, team := range teams {
			if o.MSGHandler.IsInitialised(team.ContactType) {
				return errors.Errorf("Team \"%s\" has contact type \"%s\", handler not initialised for this type - check input flags", team.Name, team.ContactType)
			}
		}
	}
	return nil
}

// GetMessageBreakdown creates a breakdown of the messages found in the pr description if the messages have changed
// since the last run
func (o *Options) GetMessageBreakdown(ctx context.Context, messages []message.Message) (string, error) {
	changed, hash, err := o.HaveMessagesChanged(ctx, messages)
	if err != nil {
		return "", err
	}
	if !changed {
		return "", nil
	}
	breakdown, err := message.GenerateBreakdown(messages, len(o.Feathers.GetAllTeamNames()))
	if err != nil {
		return "", err
	}
	return comment.AddMetadataToComment(breakdown, hash, comment.BreakdownCommentType), nil
}

// HaveMessagesChanged checks if the messages have changed since the last time the breakdown was posted to the PR
func (o *Options) HaveMessagesChanged(ctx context.Context, messages []message.Message) (bool, string, error) {
	log.Info("Checking if messages have changed")
	currentHash, err := message.GenerateHash(messages)
	if err != nil {
		return false, "", err
	}

	comments, err := o.GitServerClient.GetPRComments(ctx)
	if err != nil {
		return false, "", err
	}
	if len(comments) < 1 {
		log.Info("No comments found on PR")
		return true, currentHash, nil
	}

	var previousHash string
	for _, c := range comments {
		// Comments sorted by most recent first, so the first matching comment
		// was the last one posted by the bot
		previousHash, _ = comment.GetMetadataFromComment(*c.Body)
		if previousHash != "" {
			log.Info("Found previous hash in comment")
			break
		}
	}

	if previousHash == "" {
		log.Info("No previous hash found in comments")
		return true, currentHash, nil
	}

	if previousHash == currentHash {
		log.Info("Previous hash matches current hash, messages have not changed")
		return false, "", nil
	}
	log.Info("Previous hash does not match current hash, messages have changed")
	return true, currentHash, nil
}

// PostErrorToPR posts an error to the pull request as a comment
func (o *Options) PostErrorToPR(ctx context.Context, err error) {
	// If it's not a DryRun then we shouldn't post the error back to the pr
	if o.DryRun {
		errorMsg := fmt.Sprintf("[Peacock] Validation Failed:\n%s", err.Error())
		err = o.GitServerClient.CommentOnPR(ctx, errorMsg)
		if err != nil {
			panic(err)
		}
	}
}

// initialiseFlagsAndClients checks that all the variables required to run the command are set up correctly
// and sets up the required clients
func (o *Options) initialiseFlagsAndClients() (err error) {
	// validate flags
	if o.GitHubToken == "" {
		return errors.New("github token is required")
	}
	if o.DryRun && o.PRNumber == -1 {
		return errors.New("pr-number is required")
	}
	if o.GitServerURL == "" {
		o.GitServerURL = domain.GitHubURL
	}

	// Init git clients
	if o.Git == nil {
		o.Git = git.NewClient()
	}

	if o.RepoOwner == "" || o.RepoName == "" {
		log.Info("No one repo owner or name provided, getting from git")
		o.RepoOwner, o.RepoName, err = o.Git.GetRepoOwnerAndName("")
		if err != nil {
			return errors.Wrap(err, "failed to get repo owner and name")
		}
	}

	if o.GitServerClient == nil {
		o.GitServerClient = github.NewClient(o.RepoOwner, o.RepoName, o.GitUser, o.GitHubToken, o.PRNumber)
	}
	return nil
}
