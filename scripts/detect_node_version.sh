JSON_FILE=$1

NODE_VERSION=$(cat $JSON_FILE | jq .engines.node)
NPM_VERSION=$(cat $JSON_FILE | jq .engines.npm)

echo "Node Version:" $NODE_VERSION
echo "NPM Version:" $NPM_VERSION
