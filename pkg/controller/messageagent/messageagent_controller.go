package messageagent

import (
	"context"
	"encoding/json"
	"fmt"
	messageagentv1 "github.com/gzlj/message-agent-operator/pkg/apis/messageagent/v1"
	"github.com/gzlj/message-agent-operator/pkg/resources"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"

	//"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_messageagent")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new MessageAgent Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMessageAgent{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("messageagent-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource MessageAgent
	err = c.Watch(&source.Kind{Type: &messageagentv1.MessageAgent{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner MessageAgent
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &messageagentv1.MessageAgent{},
	})


	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &messageagentv1.MessageAgent{},
	})

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &messageagentv1.MessageAgent{},
	})
	/*//appsv1.Deployment{}
	//
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &messageagentv1.MessageAgent{},
	})*/


	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileMessageAgent implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileMessageAgent{}

// ReconcileMessageAgent reconciles a MessageAgent object
type ReconcileMessageAgent struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a MessageAgent object and makes changes based on the state read
// and what is in the MessageAgent.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMessageAgent) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling MessageAgent")
	var (
		err error
	)
	// Fetch the MessageAgent instance
	instance := &messageagentv1.MessageAgent{}
	err = r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	deploy := &appsv1.Deployment{}
	oldDeploy := &appsv1.Deployment{}
	secret := &corev1.Secret{}
	oldSecret := &corev1.Secret{}

	var secretIsChanged = false

	err = r.client.Get(context.TODO(), request.NamespacedName, oldSecret)
	if err != nil && errors.IsNotFound(err) {
		secret = resources.NewSecret(instance)
		if err = r.client.Create(context.TODO(), secret); err != nil {
			return reconcile.Result{}, err
		}
	} else if err == nil {
		dataFromCr:= resources.GetSecretDataForCr(instance)
		dataFromSecret := oldSecret.Data
		if reflect.DeepEqual(dataFromCr, dataFromSecret) {
			//return reconcile.Result{}, nil
		} else {
			secret = resources.NewSecret(instance)
			secretIsChanged = true
			if err = r.client.Update(context.TODO(), secret); err != nil {
				return reconcile.Result{}, err
			}
		}

	}

	err = r.client.Get(context.TODO(), request.NamespacedName, oldDeploy)
	if err == nil {

		// need to update deploymen?
		oldDeploymentSpec := appsv1.DeploymentSpec{}
		json.Unmarshal([]byte(oldDeploy.Annotations["spec"]), &oldDeploymentSpec)
		deploy = resources.NewDeployment(instance)

		specStrFromCr := resources.GetAnnotationSpecValue(instance)
		specStrFromDeploy := resources.GetAnnotationSpecValueFromDeploy(oldDeploy)
		if  specStrFromCr !=  specStrFromDeploy {
			//update deployment
			oldDeploy.Spec = deploy.Spec
			if err = r.client.Update(context.TODO(), oldDeploy); err != nil {
				reqLogger.Info("Failed to update deployment when Reconciling MessageAgent: ",instance.Namespace + "/" + instance.Name)
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		}
		if secretIsChanged == false {
			return reconcile.Result{}, nil
		}

		// or just delete pod of this deployment
		replicas := oldDeploy.Spec.Replicas
		zero := int32(0)
		oldDeploy.Spec.Replicas=&zero
		r.client.Update(context.TODO(), oldDeploy)
		oldDeploy.Spec.Replicas=replicas
		r.client.Update(context.TODO(), oldDeploy)
		return reconcile.Result{}, nil

	} else if ! errors.IsNotFound(err){
		fmt.Println("(lse if ! errors.IsNotFound(err)): ", err)
		return reconcile.Result{}, err
	}

	//create deployment && add spec annotation
	deploy = resources.NewDeployment(instance)
	if err = r.client.Create(context.TODO(), deploy); err != nil {
		reqLogger.Info("Failed to create deployment when Reconciling MessageAgent: ",instance.Namespace + "/" + instance.Name)
		return reconcile.Result{}, err
	}

	// Pod already exists - don't requeue
	//reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	return reconcile.Result{}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *messageagentv1.MessageAgent) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}
