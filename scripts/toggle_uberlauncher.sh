#!/bin/bash

addr=$(hyprctl clients -j | jq -r '.[] | select(.class == "uberlauncher") | .address' | head -1)
if [ -n "$addr" ]; then
    hyprctl dispatch closewindow "address:$addr"
else
    footclient --term xterm-256color --app-id uberlauncher -T Launcher -e bash -lc ~/Projects/uberlauncher/uberlauncher
fi
