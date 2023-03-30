# temporal-trivia
A trivia game built on temporal

## Setup
Set the following environment variables. These variables configure the temporal namespace, endpoint and certs. In addition since chatgpt is used, a valid chatgpt API key is also required. You can create a chatgpt API key [here](https://platform.openai.com/account/api-keys).

Client Configuration parameters
<pre>
export TEMPORAL_NAMESPACE="namespace.AccountId or namespace"
export TEMPORAL_HOST_URL="$TEMPORAL_NAMESPACE.tmprl.cloud:7233 or 127.0.0.1:7233"
export TEMPORAL_MTLS_TLS_CERT="/path/to/ca.pem"
export TEMPORAL_MTLS_TLS_KEY="/path/to/ca.key"
</pre>

Worker Configuration parameters
<pre>
export TEMPORAL_NAMESPACE="namespace.AccountId or namespace"
export TEMPORAL_HOST_URL="$TEMPORAL_NAMESPACE.tmprl.cloud:7233 or 127.0.0.1:7233"
export TEMPORAL_MTLS_TLS_CERT="/path/to/ca.pem"
export TEMPORAL_MTLS_TLS_KEY="/path/to/ca.key"
CHATGPT_API_KEY="<API KEY>"
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

## Game Scoring
Players will get two points for being the first to get a right answer and one point for getting the right answer but not being first. Final scores will revealed after the game completes.

## Game Rules
Each game can be configured with its own rules. Whoever starts the game sets the rules. This is done as a workflow input.

<pre>
Category:          "General",
NumberOfQuestions: 3,
NumberOfPlayers:   2,
QuestionTimeLimit: 60,
</pre>

## Play the game
Ensure you are exporting the environment variables.
<pre>
$ cd temporal-trivia
</pre>

### Run worker
You can also use the Docker file to build a worker image and under the yaml folder is everything needed to deploy on k8s.
<pre>
$ go run worker/main.go
</pre>

### Run Solo Version using CLI
Using the CLI you can play as a single player. 

<pre>
$ go run cli/trivia.go -s --mtls-cert /home/ktenzer/temporal/certs/ca.pem --mtls-key /home/ktenzer/temporal/certs/ca.key --temporal-endpoint temporal-trivia.xyzzy.tmprl.cloud:7233 --temporal-namespace temporal-trivia.xyzzy --questions 5 --category geography
</pre>

<pre>
What is the largest country in the world by land area? 
A) Russia 
B) China 
C) United States 
D) Canada

Answer: a
Correct Answer: A
Which country is the largest producer of coffee in the world?
A) Brazil
B) Colombia
C) Ethiopia
D) Vietnam

Answer: a
Correct Answer: A
What is the smallest country in the world by land area?
A) Monaco
B) San Marino
C) Vatican City
D) Liechtenstein

Answer: a
Correct Answer: C
Which of these African countries is NOT along the equator?
A) Democratic Republic of Congo
B) Kenya
C) Uganda
D) Tanzania

Answer: a
Correct Answer: B
Which body of water is located between Turkey and Ukraine?
A) Black Sea
B) Mediterranean Sea
C) Caspian Sea
D) Adriatic Sea

Answer: a
Correct Answer: A
***** Your Score *****
solo 6
</pre>

### Experimenting 
You can manually run a game and understand how the interaction works using the starter, signaler and query programs. First set the client environment variables.

#### Start the game (workflow)
<pre>
$ go run starter/main.go
</pre>

Each game is a workflow. Starting the workflow starts the game. We interact with the game by querying and sending signals to workflow using workflowId.

<pre>
2023/03/27 18:50:25 Started workflow WorkflowID trivia_game_152a2c56-35fc-4e0d-96e9-b5b9544ab9a9 RunID 06f87678-06b7-404b-8629-5ead6cc06e96
</pre>

### Query the game (workflow)
Query game for questions and status of progress, set the workflowId via environment from return of workflow start. There are three different queries: getDetails, getProgress and getScore. As the game progress and players answer questions, these maps will dynamically update.
<pre>
$ TEMPORAL_WORKFLOW_ID=trivia_game_152a2c56-35fc-4e0d-96e9-b5b9544ab9a9 go run query/main.go
</pre>

Since we just started the game all we see is the question and multiple choice answers and the correct answer (don't cheat).
<pre>
map[0:map[answer:C multipleChoiceAnswers:map[A:7 B:9 C:11 D:13] question:What is the maximum number of players allowed on the field for a soccer team?
A) 7
B) 9
C) 11
D) 13 submissions:map[[]] winner:]] 1:map[answer:D multipleChoiceAnswers:map[A:Sydney B:Perth C:Melbourne D:Canberra] question:What is the capital city of Australia?
A) Sydney
B) Perth
C) Melbourne
D) Canberra submissions:map[john:map[answer:B isCorrect:false] keith:map[answer:A isCorrect:false]] winner:] 2:map[answer:B multipleChoiceAnswers:map[A:J B:J C:Stephenie Meyer D:Suzanne Collins] question:Who is the author of the Harry Potter book series?
A) J.R.R. Tolkien
B) J.K. Rowling
C) Stephenie Meyer
D) Suzanne Collins submissions:map[[]] winner:]]
map[]
map[]
</pre>

### Send answers per player to game (workflow)
Using a signal players can respond to the question with their answers. The player and the answer are set via environment parameters.

ANthony answers a
<pre>
TEMPORAL_WORKFLOW_ID=trivia_game_152a2c56-35fc-4e0d-96e9-b5b9544ab9a9 TEMPORAL_TRIVIA_PLAYER=anthony TEMPORAL_TRIVIA_ANSWER=b go run signaler/main.go 
</pre>

Keith answers c
<pre>
TEMPORAL_WORKFLOW_ID=trivia_game_152a2c56-35fc-4e0d-96e9-b5b9544ab9a9 TEMPORAL_TRIVIA_PLAYER=keith TEMPORAL_TRIVIA_ANSWER=A go run signaler/main.go
</pre>

### Query the game for progress (workflow)
We can query at any time to get current status of the game and a scoreboard.
<pre>
map[0:map[answer:B multipleChoiceAnswers:map[A:Mars B:Jupiter C:Saturn D:Neptune] question:What is the name of the largest planet in our solar system? 

A) Mars
B) Jupiter
C) Saturn
D) Neptune
 submissions:map[anthony:map[answer:B isCorrect:true] keith:map[answer:A isCorrect:false]] winner:anthony] 1:map[answer:A multipleChoiceAnswers:map[A:Paris B:Venice C:Barcelona D:Rome] question:What city is famously known as the "City of Love"? 

A) Paris
B) Venice
C) Barcelona
D) Rome
 submissions:map[anthony:map[answer:B isCorrect:false] keith:map[answer:A isCorrect:true]] winner:keith] 2:map[answer:B multipleChoiceAnswers:map[A:Monaco B:Vatican City C:San Marino D:Liechtenstein] question:What is the smallest country in the world by land area? 

A) Monaco
B) Vatican City
C) San Marino
D) Liechtenstein
 submissions:map[anthony:map[answer:B isCorrect:true] keith:map[answer:A isCorrect:false]] winner:anthony]]
map[anthony:4 keith:2]
map[currentQuestion:4 numberOfQuestions:3]
</pre>

![Event History](/img/history.png)
