<!-- Provide a general summary of your changes in the Title above -->

## Description
<!-- Describe your changes in detail -->

## Related Issue
<!-- If suggesting a new feature or change, please discuss it in an issue first -->
<!-- If fixing a bug, there should be an issue describing it with steps to reproduce -->
<!-- Please link to the issue here -->

## Motivation and Context
<!-- Why is this change required? What problem does it solve? -->

## Test Coverage For This Change
<!-- Please describe in detail how you manually tested your changes, and where any automated test coverage was added/updated -->
<!-- Include details of your testing environment, and the tests you ran to see how your change affects other areas of the code, etc. -->

## Screenshots (if appropriate)

## Types of changes
<!-- What types of changes does your code introduce? Put an `x` in all the boxes that apply: -->
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to change)

## Checklist
<!-- Go over all the following points, and put an `x` in all the boxes that apply -->
<!-- If you're unsure about any of these, don't hesitate to ask; we're here to help! -->
- [ ] Make sure you are requesting to **pull a topic/feature/bugfix branch** (right side). If PRing from your fork, don't come from your `master`!
- [ ] Make sure you are making a pull request against our **`master` branch** (left side). Also, it would be best if you started *your change* off *our latest `master`*.
- [ ] Make sure you are updating [CHANGELOG.md](https://github.com/TykTechnologies/tyk-operator/blob/master/CHANGELOG.md) based on your changes.
- [ ] My change requires a change to the documentation.
  - [ ] If you've changed APIs, describe what needs to be updated in the documentation.
- [ ] I have updated the documentation accordingly.
- [ ] If you've changed API models, please update CRDs.
  - [ ] `make manifests`
  - [ ] `make helm`
- [ ] I have added tests to cover my changes.
- [ ] All new and existing tests passed.
- [ ] Check your code additions will not fail linting checks:
  - [ ] `gofmt -s -w .`
  - [ ] `go vet ./...`
  - [ ] `golangci-lint run`
