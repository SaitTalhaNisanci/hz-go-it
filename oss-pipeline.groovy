pipeline {
    agent {
        label "docker"
    }

    parameters {
        string(name: 'NAME', defaultValue: 'runner', description: 'Image name')
    }

    options {
        timestamps()
    }

    stages {

        stage('Install Docker Compose') {
            steps {
                sh "sudo yum update -y"
                sh "sudo pip install docker-compose"
                sh "docker-compose version"
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
            //TODO it can be change withRun docker plugin command/api
            steps {
                sh "docker run -v /var/run/docker.sock:/var/run/docker.sock -ti ${params.NAME}:${env.BUILD_ID}"
            }
            //TODO stop and remove running docker
        }
    }

    post {
        always {
            sh "docker-compose -f deployment.yaml down || true"
            cleanWs deleteDirs: true
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
