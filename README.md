<p align="center">
  <img alt="peacock logo" src="images/peacock.png" height="300" />
  <h1 align="center">Peacock</h1>
  <p align="center"><B>Show off with minimal effort</B></p>
</p>

`peacock` is a simple CI/CD tool for telling your users what you're up to. It integrates with existing pipelines to
fire out notifications to your users without having to collate and write release notes.

# Overview
Peacock works by parsing the contents of the description of a Pull Request and converting that into notifications to
be sent to users - making it easy for developers to communicate release information. Peacock supports sending multiple
messages to different teams allowing you to curate your release notes based on the audience.

* Easy to set up and integrate
* Supports multiple methods of communications
* Can send multiple messages to multiple teams all from one PR

# Installation
## Local
To run Peacock on your local machine:
```bash
git clone https://github.com/spring-financial-group/peacock.git
make install
```
## CI/CD

To run Peacock in a CI/CD pipeline:
```yaml
- image: mqubeoss.azurecr.io/spring-financial-group/peacock:latest
```
Checkout [our pipeline definitions](https://github.com/spring-financial-group/peacock/tree/main/.lighthouse/jenkins-x/peacock)
for an example of how it can be used.

## Configuration
### Feathers
Peacock's feathers are each way that you'd like to communicate with your users. These are stored in `.peacock/feathers.yaml`
in the repository that you'd like Peacock to run in.

Each team in the feathers needs to have a `contactType` and some `addresses` which define how your users will be contacted.
See [Communication Methods](#communication-methods) for all the methods supported by Peacock.

```yaml
teams:
  - name: QA
    contactType: slack
    addresses:
    - C56H7G209DF
  - name: FrontEnd
    contactType: webhook
    addresses:
      - john.smith@google.com
      - tom.allen@github.com
  - name: BackEnd
    contactType: slack
    addresses:
      - C56H7G209DF
  - name: Business
    contactType: slack
    addresses:
      - C56H7G209DF
```

### Environment Variables
Environment variables are used configure Peacock in a pipeline. For integrating into different CI/CD tools the keys for
these variables can be overridden using flags for each command.

| Variable         | Description                                     | Required                                                              | Overwrite Flag            |
|------------------|-------------------------------------------------|-----------------------------------------------------------------------|---------------------------|
| `PR_NUMBER`      | The number of the Pull Request.                 | Required when using `--dry-run`                                       | `pr-number-key`           |
| `GITHUB_TOKEN`   | The token to use for authentication with GitHub | Always                                                                | `git-token-key`           |
| `REPO_OWNER`     | The owner of the repository                     | If not passed then value is retrieved from the local git instance     | `git-owner-key`           |
| `REPO_NAME`      | The name of the repository                      | If not passed then value is retrieved from the local git instance     | `git-repo-key`            |
| `GIT_SERVER`     | The domain of the git server                    | Default is https://github.com                                         | `git-server-key`          |
| `SLACK_TOKEN`    | The token used to authenticate with Slack       | Only if the `slack` communication method is defined in the feathers   | `slack-token-key`         |
| `WEBHOOK_URL`    | The URL that Peacock will post to               | Only if the `webhook` communication method is defined in the feathers | `webhook-URL-key`         |
| `WEBHOOK_SECRET` | The secret used to authenticate Peacock         | Only if the `webhook` communication method is defined in the feathers | `webhook-HMAC-secret-key` |

## Communication Methods
### Slack
To use Slack as a method of communication a Slack app will need to be setup for your organisation with the minimum
scope of `chat:write`. It's important to remember that for private channels your app will need to be invited for Peacock
to post messages.

Although Peacock does convert Markdown to Slack's implementation, due to Slack only supporting a mild version of Markdown
this conversion is limited.

### Webhook
Peacock offers a webhook so that it can be intergrated with your own communication method. When a notification is sent
peacock will send an HTTP POST request to the configured webhook URL with the following JSON body:
```json
{
   "body": "string",
   "subject": "string",
   "addresses": [
      "string"
   ]
}
```
To authenticate the request Peacock sends a `X-Signature-256` header with the request. This is a HMAC digest of the request
body using the SHA-256 hash function and the webhook secret as the key.

Peacock converts the GitHub markdown to HTML before sending the request.

# Usage
Peacock uses Notify headers (`### Notify`) in the description of a PR to identify messages and teams to contact.
Additional information about the PR can be added as long as it is above the first Notify header - otherwise it will be
included in one of the messages.

Example PR description:
```markdown
# Production Release PR
Here is some text that won't be sent in any of the messages.

### Notify QA, FrontEnd, BackEnd
# Service Promotions

**Services Being Promoted**
* Peacock

**What functionality is being released?**
* A really cool dev tool that lets you communicate release notes to your users more easily

**Risk Of Release**
Low

### Notify Business
# New Software Release
We have just promoted a new tool that will let us more easily inform you of any future releases that we make.

We will be using this from now on to communicate any really cool features that we add to the platform.
```

**Pre-submission**

Use the command `peacock run --dry-run` pre-submission to validate the messages and check that all the right
information was supplied for Peacock to run. Adding the `--comment-validation` flag means that Peacock will post a breakdown of the
messages back to the PR as a comment.

The Pull Request number needs to be provided for Peacock to run pre-submission.

**Post-submission**

Use the command `peacock run` post-submission to actually send the messages to the different teams.

## User Guide
1. Open a PR in repository as normal and add a Notify header to the PR description containing the teams you would like
   to notify (comma separated). The teams you can choose from, and their method of contact, are stored in
   `.peacock/feathers.yaml` in the repository.
2. Underneath the Notify header add the content of the message you would like to send. Keep in mind that some methods
   of contact support only a limited form of markdown - Slack.
3. Once opened, the peacock dry run pipeline will run. This parses and validates the teams & messages, posting an
   explanation back to the PR if it fails.
4. Once the PR merges the peacock release pipeline starts. This is the pipeline that actually sends the notifications.

Boop2!