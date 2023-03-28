# temporal-trivia
A trivia game built on temporal

## Setup
Set the following environment variables. These variables configure the temporal cloud namespace, endpoint and certs. In addition since chatgpt is used a valid chatgpt API key is required. You can create a chatgpt API key [here](https://platform.openai.com/account/api-keys).

Configuration parameters
<pre>
export MTLS=true
export TEMPORAL_NAMESPACE="namespace.AccountId"
export TEMPORAL_HOST_URL="$TEMPORAL_NAMESPACE.tmprl.cloud:7233"
export TEMPORAL_MTLS_TLS_CERT="/path/to/ca.pem"
export TEMPORAL_MTLS_TLS_KEY="/path/to/ca.key"
export CHATGPT_API_KEY="key"
</pre>

Game parameters
<pre>
export TEMPORAL_WORKFLOW_ID="trivia_game_152a2c56-35fc-4e0d-96e9-b5b9544ab9a9"
export TEMPORAL_TRIVIA_PLAYER="Keith"
export TEMPORAL_TRIVIA_ANSWER="A"
</pre>

<pre>
$ git clone https://github.com/ktenzer/temporal-trivia.git
</pre>

## Play the game
<pre>
$ cd temporal-trivia
</pre>

### Run worker
<pre>
$ go run worker/main.go
</pre>

### Start Worfklow (game)
You can choose how many players, questions and even the time limit per question.

<pre>
$ go run starter/main.go
</pre>

Each game is a workflow. Starting the workflow starts the game. We interact with the game by querying and sending signals to workflow using workflowId.

<pre>
2023/03/27 18:50:25 Started workflow WorkflowID trivia_game_152a2c56-35fc-4e0d-96e9-b5b9544ab9a9 RunID 06f87678-06b7-404b-8629-5ead6cc06e96
</pre>

### Query Workflow
Query game for questions and status of progress, set the workflowId via environment from return of workflow start.
<pre>
$ TEMPORAL_WORKFLOW_ID=trivia_game_152a2c56-35fc-4e0d-96e9-b5b9544ab9a9 go run query/main.go
</pre>

Since we just started the game all we see is the question and multiple choice answers and the correct answer (don't cheat).
<pre>
map[0:map[answer:C multipleChoiceAnswers:<nil> question:What is the largest organ in the human body? 

A) Liver 
B) Brain 
C) Skin 
D) Heart 

Answer: C) Skin submissions:<nil> winner:]]
map[]
</pre>

### Send Answers per Player
Using a signal players can respond to the question with their answers. The player and the answer are set via environment parameters.

John answers a
<pre>
TEMPORAL_WORKFLOW_ID=trivia_game_152a2c56-35fc-4e0d-96e9-b5b9544ab9a9 TEMPORAL_TRIVIA_PLAYER=john TEMPORAL_TRIVIA_ANSWER=a go run signaler/main.go 
</pre>

Keith answers c
<pre>
TEMPORAL_WORKFLOW_ID=trivia_game_152a2c56-35fc-4e0d-96e9-b5b9544ab9a9 TEMPORAL_TRIVIA_PLAYER=keith TEMPORAL_TRIVIA_ANSWER=c go run signaler/main.go
</pre>

### Query Game for Progress
We can query at any time to get current status of the game and a scoreboard. New questions will show up until we have ran through all questions and then the workflow as well as game are completed.
<pre>
$ TEMPORAL_WORKFLOW_ID=trivia_game_152a2c56-35fc-4e0d-96e9-b5b9544ab9a9 go run query/main.go
map[0:map[answer:C multipleChoiceAnswers:map[A:Liver B:Brain C:Skin D:Heart] question:What is the largest organ in the human body? 

A) Liver 
B) Brain 
C) Skin 
D) Heart 

Answer: C) Skin submissions:map[john:map[answer:a isCorrect:false] keith:map[answer:c isCorrect:true]] winner:keith]
</pre>

![Event History](/img/history.png)
