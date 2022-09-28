# IS_DEV=true means k8s nodecount=1
function verify_IS_DEV(){
    local nodecount=$(kubectl get nodes --no-headers | wc -l)
    IS_DEV=${IS_DEV:-false}
    echo "IS_DEV: ${IS_DEV}, nodecount=${nodecount}"
    if [ $nodecount -gt 1 ]; then
        echo "gt 1"
        if ${IS_DEV}; then
            echo "Mismatch: IS_DEV is true but found k8s node count: ${nodecount}"
            exit 1
        fi
    else
        echo "less or equal 1"
        if ! ${IS_DEV}; then
            echo "Mismatch: IS_DEV is false but found k8s node count: ${nodecount}"
            exit 1
        fi
    fi
}

function is_minikube() {
   [[ $(kubectl get nodes -o jsonpath='{.items[0].metadata.name}') == "minikube" ]]
}
