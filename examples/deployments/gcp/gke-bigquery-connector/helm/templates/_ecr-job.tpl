{{- define "connector.ecrJob" }}
template:
  spec:
    serviceAccountName: formal-ecr-secret-manager
    containers:
      - name: ecr-cred-helper
        image: amazon/aws-cli:latest
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 100m
            memory: 128Mi
        command:
          - /bin/sh
          - -c
          - |
            curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
            chmod +x kubectl
            mv kubectl /usr/local/bin/
            TOKEN=$(aws ecr get-login-password --region {{ .Values.ecrCredentials.region }})
            kubectl delete secret formal-ecr-secret || true
            kubectl create secret docker-registry formal-ecr-secret \
              --docker-server={{ .Values.ecrCredentials.registryUrl }} \
              --docker-username=AWS \
              --docker-password="${TOKEN}"
        env:
          - name: AWS_ACCESS_KEY_ID
            valueFrom:
              secretKeyRef:
                name: {{ include "connector.fullname" . }}-ecr
                key: ecr-access-key-id
          - name: AWS_SECRET_ACCESS_KEY
            valueFrom:
              secretKeyRef:
                name: {{ include "connector.fullname" . }}-ecr
                key: ecr-secret-access-key
    restartPolicy: OnFailure
{{- end }}
