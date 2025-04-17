package controller

import (
	context "context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	client "sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	authv1 "github.com/GooseFuse/distributed-auth-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

var log = logf.Log.WithName("authcluster-controller")

const (
	appLabel     = "auth-node"
	volumeName   = "data"
	mountPath    = "/app/data"
	configMapKey = "peers.conf"
	port         = 6333
)

type AuthClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *AuthClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	// Fetch the AuthCluster instance
	var cluster authv1.AuthCluster
	if err := r.Get(ctx, req.NamespacedName, &cluster); err != nil {
		if apierrors.IsNotFound(err) {
			// Object not found — probably deleted
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Determine the number of replicas
	replicas := int32(cluster.Spec.NodeCount)

	// ✅ Build the StatefulSet object
	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: "headless-service", // must match headless Service
			Replicas:    &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "auth-node",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "auth-node",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "authnode",
							Image:           "distributed-auth-system:latest",
							ImagePullPolicy: corev1.PullNever,
							Env: []corev1.EnvVar{
								{
									Name: "NODE_ID",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name:  "REDIS_URL",
									Value: cluster.Spec.RedisURL,
								},
								{
									Name:  "PEER_LIST",
									Value: generatePeerList(cluster.Name, int(cluster.Spec.NodeCount)),
								},
								{
									Name:  "PORT",
									Value: "8080",
								},
							},
							Ports: []corev1.ContainerPort{
								{ContainerPort: 6333},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/app/data",
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
						Resources: corev1.VolumeResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
					},
				},
			},
		},
	}

	// Ensure cleanup happens when CR is deleted
	if err := ctrl.SetControllerReference(&cluster, statefulSet, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// ✅ Create or update the StatefulSet
	if err := r.Create(ctx, statefulSet); err != nil && !apierrors.IsAlreadyExists(err) {
		log.Error(err, "Failed to create StatefulSet")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *AuthClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&authv1.AuthCluster{}).
		Complete(r)
}

func generatePeerList(baseName string, replicas int) string {
	peers := make([]string, replicas)
	for i := 0; i < replicas; i++ {
		peers[i] = fmt.Sprintf("%s-%d.headless-service.default.svc.cluster.local:%d", baseName, i, port)
	}
	return strings.Join(peers, ",")
}
