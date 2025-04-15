package controller

import (
	context "context"
	fmt "fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	client "sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	authv1 "github.com/GooseFuse/distributed-auth-operator/api/v1"
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
	log := log.WithValues("authcluster", req.NamespacedName)
	log.Info("Reconciling AuthCluster", "name", req.Name)
	var cluster authv1.AuthCluster
	if err := r.Get(ctx, req.NamespacedName, &cluster); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Ensure peer ConfigMap exists
	peerData := ""
	for i := 0; i < cluster.Spec.NodeCount; i++ {
		peerData += fmt.Sprintf("auth-node-%d:%d\n", i, port)
	}
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-peers", cluster.Name),
			Namespace: req.Namespace,
		},
		Data: map[string]string{
			configMapKey: peerData,
		},
	}
	_ = ctrl.SetControllerReference(&cluster, cm, r.Scheme)
	r.Create(ctx, cm)

	// Create pods up to nodeCount
	for i := 0; i < cluster.Spec.NodeCount; i++ {
		podName := fmt.Sprintf("auth-node-%d", i)
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      podName,
				Namespace: req.Namespace,
				Labels: map[string]string{
					"app": appLabel,
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name:            "authnode",
					Image:           "distributed-auth-system:latest",
					ImagePullPolicy: corev1.PullNever,
					Ports: []corev1.ContainerPort{{
						ContainerPort: port,
					}},
					Env: []corev1.EnvVar{
						{Name: "NODE_ID", Value: podName},
						{Name: "REDIS_URL", Value: cluster.Spec.RedisURL},
						{Name: "PEER_LIST", ValueFrom: &corev1.EnvVarSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								Key: configMapKey,
								LocalObjectReference: corev1.LocalObjectReference{
									Name: fmt.Sprintf("%s-peers", cluster.Name),
								},
							},
						}},
					},
					VolumeMounts: []corev1.VolumeMount{{
						Name:      volumeName,
						MountPath: mountPath,
					}},
				}},
				Volumes: []corev1.Volume{
					{
						Name: volumeName,
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: fmt.Sprintf("%s-pvc-%d", cluster.Name, i),
							},
						},
					},
				},
			},
		}
		r.Create(ctx, pod)

		// Create PVC
		pvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-pvc-%d", cluster.Name, i),
				Namespace: req.Namespace,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						"storage": resource.MustParse("1Gi"),
					},
				},
			},
		}
		r.Create(ctx, pvc)
	}

	return ctrl.Result{}, nil
}

func (r *AuthClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&authv1.AuthCluster{}).
		Complete(r)
}
