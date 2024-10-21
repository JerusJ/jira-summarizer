# Jira Summarizer

Summarize a user's changed Jira issues in between a start and end date range (`MM/DD/YY`)

This will, for the specified user:

- Grab issues that have changed status transitions within the date range 
    - I.E. `To Do --> In Progress`
- Grab the user's comments in those issues (also within the date range)

## Install

`go install github.com/jerusj/jira-summarizer@latest`

## Setup

You will need the following:

- `JIRA_API_TOKEN` 
    - Get it from [here](https://id.atlassian.com/manage-profile/security/api-tokens)
- `JIRA_URL`
    - Base URL to your Jira on-prem/cloud instance 
    - I.E. `https://<NAME>.atlassian.net`
- `JIRA_EMAIL` 
    - Your user's sign-in email address to the Jira instance 
    - I.E. `foo.bar@baz.com`

## Usage

Run: `jira-summarizer --template=slack --start=10/17/2024 --end=10/18/2024`

For available templates, see the directory: `./pkg/jira/templates/`
