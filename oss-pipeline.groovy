pipeline {
    agent {
        label "docker"
    }

    options {
        timestamps()
    }

    stages {

        stage('Check Docker') {
            steps {
                //todo baris add init script
                sh "sudo pip install docker-compose"
                sh "sudo yum update -y"
                sh "sudo yum install -y golang"
            }
        }

        stage('Deploy Cluster') {
            steps {
                git changelog: false, poll: false, url: 'https://github.com/lazerion/hz-go-it.git'
                sh "docker-compose -f deployment.yaml up -d"
            }
        }

        stage('Acceptance') {
            steps {
                sh "go test"
            }
        }
    }

    post {
        always {
            sh "docker-compose -f deployment.yaml down || true"
            cleanWs deleteDirs: true
        }
        failure {
            mail to: 'baris@hazelcast.com',
                    subject: "Failed Pipeline: ${currentBuild.fullDisplayName}",
                    body: "Something is wrong with ${env.BUILD_URL}"
        }
    }
}
