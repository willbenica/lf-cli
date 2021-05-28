# lf-CLI

A command line tool to get data from the [leadfeeder API](https://docs.leadfeeder.com/api/#introduction). This data is returned in JSON format and this tool paris well with `jq` for discovery.

The leadfeeder API offers various endpoints, however this tool only gets data from:

* Leads: e.g. `https://api.leadfeeder.com/accounts/<account_id>/leads?start_date=<date>&end_date=<date>`
* Visits: e.g. `GET https://api.leadfeeder.com/accounts/<account_id>/visits?start_date=<date>&end_date=<date>`

When called from the CLI, you have the choice to either get only one page and print to standard out, or get all leads which will make multiple calls (if necessary) and save the response in a file(s).

## Installation:

Below are the steps to install (tested only with MacOS BigSur)

### Requirements

* Go version 1.15+
* Leadfeeder account ID (6-digit number)
* Leadfeeder API token

## Creating a lf-cli.yaml configuration file

The file should be located under `$HOME/.conf/lf-cli/.lf-cli.yaml` or `$HOME/.lf-cli.yaml`
Contents:

```yaml
lf-url: "my.leadfeeder.me"
account: "123456"
token:  "xxxxYYYYxxxxWWWWxxxxQQQ867512"
```

## Example usage:

* Print the `lf-cli` help page

    ```zsh
    $ lf-cli help
    Get leadfeeder data from a specific API endpoint and push to a local file (JSON).
    For ease of use create a config file under $HOME/.config/lf-cli/.lf-cli.yaml
    or under $HOME/.lf-cli.yaml with the following
      lf-url:  "my.leadfeeder.me"
      account: "myAccountID"
      token:   "myApiToken"

    Usage:
      lf-cli [command]

    Available Commands:
      get         Get the data from an endpoint, e.g. 'leads', 'custom-feeds', etc
      help        Help about any command

    Flags:
          --config string      path to a config file (default is $HOME/.config/lf-cli/.lf-cli.yaml)
          --lf-url string      leadfeeder URL (default "https://api.leadfeeder.com")
          --accountID string   Account for which data should be accessed
          --token string       API token used to access lf
      -v, --verbose            Increases loglevel to DEBUG for trouble shooting.
      -h, --help               help for lf-cli

    Use "lf-cli [command] --help" for more information about a command.
    ```

* Get one the 100th page of leads with 100 leads per page without a config file:

    ```zsh
    lf-cli get leads --accountID 123456 --lf-url "api.leadfeeder.com" --token "xxxxYYYYxxxxWWWWxxxxQQQ867512" -n 100 -z 100
    ```

* Get a page the fist page of leads with 25 leads per page and pipe to `jq`

    ```zsh
    lf-cli get leads -z 25 -n 1 -s 2021-01-01 | jq .
    ```

### Using `lf-cli` with `jq`

```zsh
❯ lf-cli get visits -s 2021-05-27 | jq '[.data[].attributes | {id: .lead_id, duration: .visit_length, campaign} | select(.campaign!=null)] | group_by(.campaign)| [.[] | {campaign: .[0].campaign, count: . | length}]| sort_by(.count) | reverse'
[
  {
    "campaign": "Campaign_1",
    "count": 67
  },
  {
    "campaign": "Campaign_2",
    "count": 28
  },
  {
    "campaign": "Campaign_3",
    "count": 3
  },
  {
    "campaign": "Campaign_4",
    "count": 1
  },
  {
    "campaign": "Campaign_5",
    "count": 1
  }
]
```
