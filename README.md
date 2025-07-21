# gh-fuda

`gh-fuda` is a gh extension which extends label manipulations.

## Installation

Install as a [gh](https://cli.github.com/) extension (ref. [gh manual of gh extension install](https://cli.github.com/manual/gh_extension_install)).

```bash
gh extension install tnagatomi/gh-fuda
```

## Usage

Login to GitHub with `gh auth login` (ref. [gh manual of gh auth login](https://cli.github.com/manual/gh_auth_login)) so that the extension can access the repositories.

### Global Options

- `-R`, `--repos`: Select repositories using the `OWNER/REPO` format separated by comma (e.g., `owner1/repo1,owner2/repo1`)
- `--dry-run`: Check what operations would be executed without actually operating on the repositories

### List of Commands

#### Create Labels

```bash
gh fuda create
```

Create specified labels to the specified repositories.

##### Options

- `-l`, `--labels`: Specify the labels to create in the format of `'label1:color1:description1[,label2:color2:description2,...]'` (description can be omitted)
- `--json`: Specify the path to a JSON file containing labels to create

##### Example

```bash
# Using inline labels
gh fuda create -R "owner1/repo1,owner1/repo2,owner2/repo1" -l "label1:ff0000:description for label 1,label2:00ff00,label3:0000ff"

# Using JSON file
gh fuda create -R "owner1/repo1,owner1/repo2,owner2/repo1" --json labels.json
```

##### JSON File Format

```json
[
  {
    "name": "bug",
    "color": "d73a4a",
    "description": "Something isn't working"
  },
  {
    "name": "enhancement",
    "color": "a2eeef",
    "description": "New feature or request"
  },
  {
    "name": "documentation",
    "color": "0075ca",
    "description": "Improvements or additions to documentation"
  }
]
```

#### Delete Labels

```bash
gh fuda delete
```

Delete specified labels from the specified repositories.

##### Options

- `-l`, `--labels`: Specify the labels to delete in the format of `'label1[,label2,...]'`
- `--force`: Do not prompt for confirmation

##### Example

```bash
gh fuda delete -R "owner1/repo1,owner1/repo2,owner2/repo1" -l "label1,label2,label3"
```

#### Sync Labels

```bash
gh fuda sync
```

Sync the labels in the specified repositories with the specified labels.

##### Options

- `-l`, `--labels`: Specify the labels to set in the format of `'label1:color1:description1[,label2:color2:description2,...]'` (description can be omitted)
- `--json`: Specify the path to a JSON file containing labels to sync
- `--force`: Do not prompt for confirmation

##### Example

```bash
# Using inline labels
gh fuda sync -R "owner1/repo1,owner1/repo2,owner2/repo1" -l "label1:ff0000:description for label 1,label2:00ff00,label3:0000ff"

# Using JSON file
gh fuda sync -R "owner1/repo1,owner1/repo2,owner2/repo1" --json labels.json
```

##### JSON File Format

The JSON file format is the same as the one used for the `create` command. See the [Create Labels](#create-labels) section for details.

#### Empty Labels

```bash
gh fuda empty
```

Delete all labels from the specified repositories.

##### Options

- `--force`: Do not prompt for confirmation

##### Example

```bash
gh fuda empty -R "owner1/repo1,owner1/repo2,owner2/repo1"
```
