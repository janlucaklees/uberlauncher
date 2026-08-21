#!/bin/bash

addr=$(hyprctl clients -j | jq -r '.[] | select(.class == "uberlauncher") | .address' | head -1)
if [ -n "$addr" ]; then
    hyprctl dispatch "hl.dsp.window.close(\"address:$addr\")"
else
    footclient --term xterm-256color --app-id uberlauncher -T Launcher -e zsh -c "${HOME}/go/bin/uberlauncher"
fi
