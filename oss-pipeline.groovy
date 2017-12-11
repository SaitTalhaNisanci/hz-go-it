pipeline {
    agent {
        label "docker-image-test"
    }

    parameters {
        string(name: 'NAME', defaultValue: 'runner', description: 'Image name')
    }

    stages {

        stage('Install Tools') {
            steps {
                sh "sudo pip install docker-compose"
                sh "sudo yum install -y golang"
            }
        }

        stage('Checkout') {
            steps {
                git branch: 'master', credentialsId: '7df9a580-f2f9-4a4a-9523-b22157b6d32f', url: 'https://github.com/hazelcast/go-client.git'
                sh "mkdir -p /home/ec2-user/go/src/github.com/hazelcast/go-client"
                sh "sudo cp -R ./ /home/ec2-user/go/src/github.com/hazelcast/go-client/"
            }
        }

        stage('Load Dependencies') {
            steps {
                sh """
                        go get golang.org/x/net/context
                        go get github.com/docker/libcompose/docker
                        go get github.com/docker/libcompose/docker/ctx
                        go get github.com/docker/libcompose/project
                        go get github.com/docker/libcompose/project/options 
                        go get github.com/stretchr/testify/assert
                        go get github.com/lucasjones/reggen
                        go get github.com/montanaflynn/stats
                """
            }
        }
        stage('Build Runner') {
            steps {
                git changelog: false, poll: false, url: 'https://github.com/lazerion/hz-go-it.git'
                script {
                    runner = docker.build("${params.NAME}:${env.BUILD_ID}")
                }
            }
        }

        stage('Acceptance') {
            steps {
                sh "docker network create --attachable=true go-it"
                sh "docker run --network=go-it --name=go-it -v /home/ec2-user/go:/go -v /var/run/docker.sock:/var/run/docker.sock ${params.NAME}:${env.BUILD_ID}"
                sh "docker network rm go-it"
            }
        }
    }

    post {
        always {
            sh "docker-compose -f ./acceptance/deployment.yaml down || true"
            sh "docker stop go-it || true"
            sh "docker rm go-it || true"
            script {
                sh "docker rmi ${runner.id}"
            }
        }
        failure {
            mail to: 'baris@hazelcast.com',
                    subject: "Failed Pipeline: ${currentBuild.fullDisplayName}",
                    body: "Something is wrong with ${env.BUILD_URL}"
        }
    }
}
