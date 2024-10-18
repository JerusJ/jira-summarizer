# Jira Summarizer

Summarize Jira issues in between a start and end date range (`MM/DD/YY`)

## Install

`go install github.com/jerusj/jira-summarizer@latest`

## Setup

You will need the following:

- `JIRA_API_TOKEN`: get it from [here](https://id.atlassian.com/manage-profile/security/api-tokens)
- `JIRA_URL`: base URL to your Jira on-prem/cloud instance 
    - I.E. `https://<NAME>.atlassian.net`
- `JIRA_EMAIL`: your sign-in email address to the jira instance 
    - I.E. `foo.bar@baz.com`

## Usage

Run: `jira-summarizer --template=slack --start=10/17/2024 --end=10/18/2024`

For available templates, see the directory: `./pkg/jira/templates/`
