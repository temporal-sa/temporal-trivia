# temporal-trivia
A trivia game built on temporal

## Setup
Set the following environment variables. These variables configure the temporal cloud namespace, endpoint and certs. In addition since chatgpt is used a valid chatgpt API key is required. YOu can create a chatgpt API key [here](https://platform.openai.com/account/api-keys).

export MTLS=true
export TEMPORAL_NAMESPACE="namespace.AccountId"
export TEMPORAL_HOST_URL="$TEMPORAL_NAMESPACE.tmprl.cloud:7233"
export TEMPORAL_MTLS_TLS_CERT="/path/to/ca.pem"
export TEMPORAL_MTLS_TLS_KEY="/path/to/ca.key"
export CHATGPT_API_KEY="key"

<pre>
$ git clone https://github.com/ktenzer/temporal-trivia.git
</pre>

## Play the game
<pre>
$ cd temporal-trivia
</pre>

Run worker
<pre>
$ go run worker/main.go
</pre>

Start the game
<pre>
$ go run starter/main.go
</pre>

Query game for questions and status of progress, you need to update workflowId with the id returned from start
<pre>
$ go run query/main.go
</pre>

Send answers from players via signal, you need to update workflowId and also answers
<pre>
$ go run signaler/main.go
</pre>
