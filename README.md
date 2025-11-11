# About
This repository contains the project code for MMIV (2004), a simple one-board imageboard platform made in Go.

The project was made as practice to see how the general process of creating a full-stack application from the ground-up could be done.
Initially started out as a 'what-if', this project was aimed with the goal of having minimal dependencies and as such has a relatively 
light and fast setup.

## Installation
1. Download the repository source code from this repository.
2. Run with `go run main.go`

## Features
* GIF, JPG, JPEG, PNG upload.
* Creating posts and creating replies.
* Back-end and front-end error handling.
* Basic moderation tools.

## Cons
* Does not include rate limiting.
* Has not been tested much, no unit / integration tests.
* Not built with efficient scalability in mind.