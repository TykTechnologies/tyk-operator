# Table of Contents
1. [Contributing to Tyk Operator](#contributing-to-tyk-operator)
2. [Filling an issue](#filling-an-issue)
3. [Guidelines for Pull Requests](#guidelines-for-pull-requests)
4. [Project Structure](#project-structure)
5. [Building and Running test](#building-and-running-test)
6. [Coding Conventions](#coding-conventions)
7. [Resources](#resources)

# Contributing to Tyk Operator

**First**: if you're unsure or afraid of anything, just ask or submit an issue or pull request anyway. 
You won't be yelled at for giving your best effort. The worst that can happen is that you'll be politely asked to change
something. 
We appreciate any sort of contributions, and don't want a wall of rules to get in the way of that.

However, for those individuals who want a bit more guidance on the best way to contribute to the project, read on. 
This document will cover what we're looking for. 
By addressing all the points we're looking for, it raises the chances we can quickly merge or 
address your contributions.

### Our SLA for issues and bugs
We do value the time each contributor spends contributing to this repo, and we work hard to make sure we respond to 
your issues and Pull request as soon as we can.

Below we have outlined.

### Filling an issue
Before opening an issue, if you have a question about Tyk Operator or have a problem using it, please
start with the GitHub search and our [community forum](https://community.tyk.io).
If that doesn't answer your questions, and you have an idea for a new capability or if you think you found a bug, 
[file an issue].

### Guidelines for Pull Requests
We have created a few guidelines to help with creating PR. To make sure these requirements are followed we added 
them to the PR form as well:

1. When working on an existing issue, simply respond to the issue and express interest in working on it. 
This helps other people know that the issue is active, and hopefully prevents duplicated efforts.
2. For new ideas or breaking changes, it is always better to open an issue and discuss your idea with our team first 
before implementing it.
3. Create small Pull request that address a single issue instead of multiple issues at the same time. 
This will make it possible for the PRs to be reviewed independently.
5. Make sure to run tests locally before submitting a pull request and verify that all of them are passing.
6. Tips for making sure we review your pull request faster :
    1. Code is documented. Please comment the code where possible.
    2. Use meaningful commit messages.
    3. Keep your pull request up to date with upstream master to avoid merge conflicts.
    4. Provide a good PR description as a record of what change is being made and why it was made. 
    5. Link to a GitHub issue if it exists.
    6. Tick all the relevant checkboxes in the PR form

### Project Structure

Tyk Operator uses `Operator SDK` and uses the same project structure 
as described in `Operator SDK` [documentation](https://sdk.operatorframework.io/docs/overview/project-layout/).

### Building and Running test

#### Local Development

You can follow our [Development guideline](./docs/development.md).

### Coding Conventions
- Please make sure that your code is linted by using `make linters` command.
- If you update Custom Resource Definitions, please make sure that you generated
latest manifests, by using `make generate manifests helm`.
  - Also, make sure that your changes align with https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md

### Resources
- [How to Contribute to Open Source](https://opensource.guide/how-to-contribute/)
- [Using Pull Requests](https://help.github.com/articles/about-pull-requests/)


[file an issue]: https://github.com/TykTechnologies/tyk-operator/issues/new/choose
