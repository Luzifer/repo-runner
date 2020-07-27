# 0.17.0 / 2020-07-27

  * Rewrite presentation logic
  * Remove Windows support
  * Bundle external assets instead of direct-loading them
  * Update Dockerfile
  * Migrate to go modules setup

# 0.16.0 / 2019-05-14

  * Add config option to limit used CPU power
  * Update reporunner image

# 0.15.1 / 2019-05-14

  * Fix: Specify docker configuration directory

# 0.15.0 / 2018-09-29

  * Push ref into container

# 0.14.1 / 2018-08-22

  * Fix: Only print UUID error on errors and break on fatal

# 0.14.0 / 2018-08-22

  * Update vendors, switch to dep for vendoring
  * Fetch the runner-file for the pushed revision instead of master

# 0.13.0 / 2018-08-22

  * Add status handler for health checking
  * Remove Dockerfile (moved to separate repo)

# 0.12.0 / 2018-04-06

  * Use generic builder
  * Use global vendoring
  * Move Dockerfile to root, add license

# 0.11.0 / 2017-02-25

  * Implement auto-scrolling
  * Move to new builder image

# 0.10.0 / 2016-12-19

  * Move repo-runner to its own Github organization

# 0.9.0 / 2016-12-14

  * Make navbar stick to top

# 0.8.1 / 2016-12-14

  * Fix: Javascript switch requires break

# 0.8.0 / 2016-12-14

  * Add navbar with build details to logs page

# 0.7.0 / 2016-11-21

  * Install bash in container

# 0.6.0 / 2016-10-09

  * Fix: Add missing slash
  * Add Dockerfile
  * Allow to configure BaseURL through ENV
  * Add README
  * Support reading require-secret from env

# 0.5.1 / 2016-09-25

  * Fix: Update godeps

# 0.5.0 / 2016-09-25

  * [runner] Implement log streaming with linked URL on Github
  * [runner] Differentiate between files and volumes when mounting paths
  * [runner] Add privileged flag
  * [image] Install docker in image

# 0.4.2 / 2016-09-25

  * Fix: Forbid using CGO

# 0.4.1 / 2016-09-25

  * Remove workdir statement

# 0.4.0 / 2016-09-25

  * Make checkout dir configurable

# 0.3.1 / 2016-09-25

  * Move code into GOPATH on CI

# 0.3.0 / 2016-09-25

  * Limit builds to heads by default ignoring tags to prevent double builds when pushing tag and head

# 0.2.1 / 2016-09-25

  * Fix: Install linters before using

# 0.2.0 / 2016-09-25

  * Improve code quality by following linter advices
  * Publish to Github releases

# 0.1.0 / 2016-09-24

  * Initial version
