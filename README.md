# About
This repository contains the project code for MMIV (2004), a simple one-board imageboard platform made in Go.

The project was made as practice to see how the general process of creating a full-stack application from the ground-up could be done.
Initially started out as a 'what-if', this project was aimed with the goal of having minimal dependencies and as such has a relatively 
light and fast setup.

## Local Installation
1. Download the repository source code from this repository.
2. Run with `go run main.go`

## Features
* Posting
     * Account-based posting
     * IDs, timestamps, total replies
     * Optional hidden username posting
     * Optional PNG, GIF, JPEG uploads
* Moderating
     * Pinning, locking and deletion of user posts

## Cons
* Does not include rate limiting and safety features
* Has not been tested much, no unit / integration tests
* Not built with efficient scalability in mind

## Sample Login Accounts
* sannu -> admin
* newUser -> password