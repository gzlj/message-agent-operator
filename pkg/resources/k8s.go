package resources

import (
	"encoding/json"
	messageagentv1 "github.com/gzlj/message-agent-operator/pkg/apis/messageagent/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strconv"
)

func NewSecret(agent *messageagentv1.MessageAgent) *corev1.Secret {


	var (
		channelsStr string
		receiversStr string
		bytes []byte
	)
	for _, ch := range agent.Spec.Channels {
		channelsStr += ch + ","
	}
	channelsStr = channelsStr[0:len(channelsStr)-1]

	//RECEIVERS
	bytes, _ = json.Marshal(agent.Spec.Receivers)
	receiversStr = string(bytes)

	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      agent.Name,
			Namespace: agent.Namespace,
			Labels: map[string]string{
				"alertmanager": "receiver",
				"type": "message-center",
			},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(agent, schema.GroupVersionKind{
					Group:   messageagentv1.SchemeGroupVersion.Group,
					Version: messageagentv1.SchemeGroupVersion.Version,
					Kind:    "MessageAgent",
				}),
			},
		},
		Data: map[string][]byte{
			"messageCenter": []byte(agent.Spec.MessageCenter),
			"clientId":      []byte(agent.Spec.ClientId),
			"clientSecret":  []byte(agent.Spec.ClientSecret),
			"serverPort":    []byte(agent.Spec.ServerPort),
			"applyMsgType":  []byte(agent.Spec.ApplyMsgType),
			"channels":      []byte(channelsStr),
			"receivers":     []byte(receiversStr),
		},
		Type: "Opaque",
	}
}

func NewDeployment(agent *messageagentv1.MessageAgent) *appsv1.Deployment {
	var (
		data []byte
	)
	labels := map[string]string{"app": agent.Name}
	selector := &metav1.LabelSelector{MatchLabels: labels}
	/*

	 */
	spec := appsv1.DeploymentSpec{
		Replicas: agent.Spec.Size,
		Selector: selector,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				Containers: buildContainers(agent),
			},
		},
	}
	deploy := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",

		},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
			Name:      agent.Name,
			Namespace: agent.Namespace,
			Labels: map[string]string{
				"alertmanager": "receiver",
				"type": "message-center",
				"receiver": agent.Name,
			},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(agent, schema.GroupVersionKind{
					Group:   messageagentv1.SchemeGroupVersion.Group,
					Version: messageagentv1.SchemeGroupVersion.Version,
					Kind:    "MessageAgent",
				}),
			},
		},
		Spec: spec,
	}
	data, _ = json.Marshal(spec)
	deploy.Annotations["spec"] = string(data)

	return &deploy
}

func GetAnnotationSpecValue(agent *messageagentv1.MessageAgent) string {

	labels := map[string]string{"app": agent.Name}
	selector := &metav1.LabelSelector{MatchLabels: labels}
	spec := appsv1.DeploymentSpec{
		Replicas: agent.Spec.Size,
		Selector: selector,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				Containers: buildContainers(agent),
			},
		},
	}
	data, _ := json.Marshal(spec)
	return string(data)
}

func GetAnnotationSpecValueFromDeploy(d *appsv1.Deployment) string{
	labels := map[string]string{"app": d.Name}
	selector := &metav1.LabelSelector{MatchLabels: labels}
	spec := appsv1.DeploymentSpec{
		Replicas: d.Spec.Replicas,
		Selector: selector,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				Containers: buildContainersFromDeploy(d),
			},
		},
	}
	data, _ := json.Marshal(spec)
	return string(data)

}

func buildContainersFromDeploy(d *appsv1.Deployment) []corev1.Container{
	container := d.Spec.Template.Spec.Containers[0]
	containerPorts := []corev1.ContainerPort{
		{
			Name:          container.Name,
			ContainerPort: container.Ports[0].ContainerPort,
		},
	}
	return []corev1.Container{
		{
			Name:  container.Name,
			Image: container.Image,
			//Resources:       app.Spec.Resources,
			Ports:           containerPorts,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Env: container.Env,
				},
		}


}



func buildContainers(agent *messageagentv1.MessageAgent) []corev1.Container {
	var (
		err  error
		port int
	)
	port, err = strconv.Atoi(agent.Spec.ServerPort)
	if err != nil {
		port = 8080
	}
	containerPorts := []corev1.ContainerPort{
		corev1.ContainerPort{
			Name:          agent.Name,
			ContainerPort: int32(port),
		},
	}
	return []corev1.Container{
		{
			Name:  agent.Name,
			Image: agent.Spec.Image,
			Ports:           containerPorts,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Env: []corev1.EnvVar{
				{
				Name: "MESSAGECENTER",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: agent.Name},
						Key: "messageCenter",}},
				},
				{
					Name: "CLIENTSECRET",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: agent.Name},
							Key: "clientSecret",}},
				},
				{
					Name: "CLIENTID",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: agent.Name},
							Key: "clientId",}},
				},
				{
					Name: "APPLYMSGTYPE",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: agent.Name},
							Key: "applyMsgType",}},
				},
				{
					Name: "SERVERPORT",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: agent.Name},
							Key: "serverPort",}},
				},
				{
					Name: "CHANNELS",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: agent.Name},
							Key: "channels",}},
				},
				{
					Name: "RECEIVERS",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: agent.Name},
							Key: "receivers",}},
				},


			},
		},
	}
}

func GetSecretDataForCr(agent *messageagentv1.MessageAgent) map[string][]byte {

	var (
		channelsStr string
		receiversStr string
		bytes []byte

	)
	//fmt.Println("ch := range agent.Spec.Channels: ")
	for _, ch := range agent.Spec.Channels {
		channelsStr += ch + ","
	}
	channelsStr = channelsStr[0:len(channelsStr)-1]
	bytes, _ = json.Marshal(agent.Spec.Receivers)
	receiversStr = string(bytes)

	dataMap := map[string][]byte{
		"messageCenter": []byte(agent.Spec.MessageCenter),
		"clientId":      []byte(agent.Spec.ClientId),
		"clientSecret":  []byte(agent.Spec.ClientSecret),
		"serverPort":    []byte(agent.Spec.ServerPort),
		"applyMsgType":  []byte(agent.Spec.ApplyMsgType),
		"channels":      []byte(channelsStr),
		"receivers":     []byte(receiversStr),
	}
	return dataMap
}

func NewService(agent *messageagentv1.MessageAgent) *corev1.Service {
	var (
		port int
		servicePort corev1.ServicePort
		err error
	)
	port, err = strconv.Atoi(agent.Spec.ServerPort)
	if err != nil {
		port = 8080
	}
	servicePort = corev1.ServicePort{
		Name: agent.Name,
		Port: int32(port),
	}

	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: agent.Namespace,
			Name: agent.Name,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(agent, schema.GroupVersionKind{
					Group:   messageagentv1.SchemeGroupVersion.Group,
					Version: messageagentv1.SchemeGroupVersion.Version,
					Kind:    "MessageAgent",
				}),
			},

		},
		Spec: corev1.ServiceSpec{
			Type:  corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{servicePort},
			Selector: map[string]string{
				"app": agent.Name,
			},
		},
	}
}