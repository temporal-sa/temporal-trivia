# temporal-trivia
A trivia game built on temporal

## Workflow Design
Temporal trivia utlizes two workflows. One workflow orchestrates the game and maintains game state. The other workflow adds a player to the game.
![Workflow Design](/img/Temporal_Trivia_Workflow_Design.png)

## Deploy in Kubernetes
Build docker file
<pre>
$ cd temporal-trivia
$ docker build -t ktenzer/temporal-trivia:v1.0 .
</pre>

Push docker file
<pre>
$ docker push ktenzer/temporal-trivia:v1.0
</pre>

Add SSL Certs to Secret
<pre>
$ kubectl create secret tls temporal-trivia-tls-secret --key /home/ktenzer/temporal/certs/ca.key --cert /home/ktenzer/temporal/certs/ca.pem -n temporal-trivia
</pre>

Add ChatGPT Key to Secret
<pre>
$ kubectl create secret generic chatgpt-key --from-literal=KEY=chatgptkey -n temporal-trivia
</pre>

Create Deployment (update the environment parameters before deploying)
<pre>
$ kubectl create -f yaml/deployment.yaml
</pre>

Check Pods
<pre>
$ kubectl get pods -n temporal-trivia
NAME                              READY   STATUS    RESTARTS   AGE
temporal-trivia-95b98d7d4-5mfj4   1/1     Running   0          23h
temporal-trivia-95b98d7d4-gscgm   1/1     Running   0          23h
temporal-trivia-95b98d7d4-mhb7d   1/1     Running   0          23h
</pre>

## Deploy in Local Environment
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
Each game can be configured with its own rules. Whoever starts the game sets the rules. This is done as a workflow input. If no input is provided the defaults are assumed.

<pre>
Category:          "General",
NumberOfQuestions: 5,
NumberOfPlayers:   2,
QuestionTimeLimit: 60,
ResultTimeLimit: 10,
StartTimeLimit: 300,
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

### Run Simulation
You can manually run a game and understand how the interaction works using the simulator program. The simulator will load defaults and run a game. This is good option for testing or quick demonstration. Each player will answer questions randomly and game progress will be shown.
<pre>
$ go run simulator/main.go
</pre>

### Example Event History
![Event History](/img/history.png)
