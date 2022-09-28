#!/usr/bin/env bash

# Based on Bobby theme.

SCM_THEME_PROMPT_DIRTY=" ${red}✗"
SCM_THEME_PROMPT_CLEAN=" ${bold_green}✓"
SCM_THEME_PROMPT_PREFIX=" |"
SCM_THEME_PROMPT_SUFFIX="${green}|"

GIT_THEME_PROMPT_DIRTY=" ${red}✗"
GIT_THEME_PROMPT_CLEAN=" ${bold_green}✓"
GIT_THEME_PROMPT_PREFIX=" ${green}|"
GIT_THEME_PROMPT_SUFFIX="${green}|"

RVM_THEME_PROMPT_PREFIX="|"
RVM_THEME_PROMPT_SUFFIX="|"

function kube_context() {
    context=$(kubectl config current-context | sed -e 's/gke_gp-kubernetes_us-central1-[a|f]_/gke-/')
    if [[ "$DOCKER_CERT_PATH" = *minikube* ]] ; then
        printf "${cyan}\U1F433\U2192" # SPOUTING WHALE, RIGHTWARDS ARROW
        if [[ "$context" != minikube ]] ; then
            printf "minikube "
        fi
    fi
    printf "${cyan}\U2388 $context" # HELM SYMBOL
}

function prompt_command() {
    PS1="\n$(kube_context) | ${purple}\h ${reset_color}in ${green}\w\n${bold_cyan}$(scm_char)${green}$(scm_prompt_info) ${green}→${reset_color} "
}

PROMPT_COMMAND=prompt_command;
