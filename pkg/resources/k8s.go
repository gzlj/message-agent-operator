package resources

import (
	"encoding/base64"
	"fmt"
	messageagentv1 "github.com/gzlj/message-agent-operator/pkg/apis/messageagent/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strconv"
)

func NewSecret(agent *messageagentv1.MessageAgent) *corev1.Secret {

	//b64.StdEncoding.
	var (
		messagCcenterBytes []byte
		/*
		clientId
clientSecret
serverPort
applyMsgType
channels
		 */
		clientIdBytes []byte
		clientSecretBytes []byte
		serverPortBytes []byte
		applyMsgTypeBytes []byte
		channelsBytes []byte
		channelsStr string
	)
	fmt.Println("agent.Spec.MessagCcenter: ",agent.Spec.MessagCcenter)
	fmt.Println("agent.Spec.ClientId: ",agent.Spec.ClientId)
	fmt.Println("agent.Spec.ClientSecret: ",agent.Spec.ClientSecret)
	fmt.Println("agent.Spec.ServerPort: ",agent.Spec.ServerPort)
	base64.StdEncoding.Encode(messagCcenterBytes, []byte(agent.Spec.MessagCcenter))
	fmt.Println("messagCcenterBytes", string(messagCcenterBytes))
	base64.StdEncoding.Encode(clientIdBytes, []byte(agent.Spec.ClientId))
	base64.StdEncoding.Encode(clientSecretBytes, []byte(agent.Spec.ClientSecret))
	base64.StdEncoding.Encode(serverPortBytes, []byte(agent.Spec.ServerPort))
	base64.StdEncoding.Encode(applyMsgTypeBytes, []byte(agent.Spec.ApplyMsgType))
	fmt.Println("ch := range agent.Spec.Channels: ")
	for _, ch := range agent.Spec.Channels {
		channelsStr += ch + ","
	}
	channelsStr=channelsStr[0:len(channelsStr)-1]
	base64.StdEncoding.Encode(channelsBytes, []byte(channelsStr))

	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      agent.Name,
			Namespace: agent.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(agent, schema.GroupVersionKind{
					Group:   messageagentv1.SchemeGroupVersion.Group,
					Version: messageagentv1.SchemeGroupVersion.Version,
					Kind:    "MessageAgent",
				}),
			},
		},
		Data: map[string][]byte{
			"messagCcenter": messagCcenterBytes,
			"clientId": clientIdBytes,
			"clientSecret": clientSecretBytes,
			"serverPort": serverPortBytes,
			"applyMsgType": applyMsgTypeBytes,
			"channels": channelsBytes,
		},
		Type: "Opaque",
	}
}

func NewDeployment(agent *messageagentv1.MessageAgent) *appsv1.Deployment {
	labels := map[string]string{"app": agent.Name}
	selector := &metav1.LabelSelector{MatchLabels: labels}

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind: "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:  agent.Name,
			Namespace: agent.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(agent,schema.GroupVersionKind{
					Group: messageagentv1.SchemeGroupVersion.Group,
					Version: messageagentv1.SchemeGroupVersion.Version,
					Kind: "MessageAgent",
				}),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: agent.Spec.Size,
			Selector: selector,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers:buildContainers(agent),
				},
			},
		},

	}
}

func buildContainers(agent *messageagentv1.MessageAgent) []corev1.Container {
	var (
		err error
		port int
	)
	port, err = strconv.Atoi(agent.Spec.ServerPort)
	if err != nil {
		port = 8080
	}
	containerPorts := []corev1.ContainerPort{
		corev1.ContainerPort{
			Name: agent.Name,
			ContainerPort: int32(port),
		},
	}
	return []corev1.Container{
		{
			Name:            agent.Name,
			Image:           agent.Spec.Image,
			//Resources:       app.Spec.Resources,
			Ports:           containerPorts,
			ImagePullPolicy: corev1.PullIfNotPresent,
		},
	}
}