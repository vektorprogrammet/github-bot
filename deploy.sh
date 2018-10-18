#!/usr/bin/env bash
GOOS=linux GOARC=amd64 go build -o github-bot
if [ $? -eq 0 ]; then
    ssh vektorprogrammet@146.185.138.230 'sudo service github-bot stop'
    scp github-bot vektorprogrammet@146.185.138.230:/var/www/github-bot
    ssh vektorprogrammet@146.185.138.230 'sudo service github-bot start'
else
    echo "Build was unsuccessful"
fi
