package nfs

import (
	"log"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (d *Deployment) createStatefulSet(size resource., nfsPort int, rpcPort int) error {

	// ss := &appsv1.StatefulSet{}
	replicas := int32(1)

	ss := &appsv1.StatefulSet{
		// TODO - TypeMeta not needed?
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.nfsServer.Name,
			Namespace: d.nfsServer.Namespace,
			// Labels:          labelsForStatefulSet(d.nfsServer.Name),
			OwnerReferences: d.nfsServer.ObjectMeta.OwnerReferences,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: d.nfsServer.Name,
			Replicas:    &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labelsForStatefulSet(d.nfsServer.Name),
			},
			Template:             d.createPodTemplateSpec(nfsPort, rpcPort),
			VolumeClaimTemplates: d.createVolumeClaimTemplateSpecs(size),
		},
	}

	log.Printf("ss: %#v", ss)

	// podSpec := &sset.Spec.Template.Spec

	// s.addPodPriorityClass(podSpec)

	// s.addNodeAffinity(podSpec)

	// if err := s.addTolerations(podSpec); err != nil {
	// 	return err
	// }

	return d.createOrUpdateObject(ss)
}

func (d *Deployment) createVolumeClaimTemplateSpecs(size resource.Quantity) []corev1.PersistentVolumeClaim {

	scName := "fast"

	return []corev1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				// Name:      d.nfsServer.Name,
				Name:      "nfs-data",
				Namespace: d.nfsServer.Namespace,
				Labels:    labelsForStatefulSet(d.nfsServer.Name),
				Annotations: map[string]string{
					"volume.beta.kubernetes.io/storage-class": "fast",
				},
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				StorageClassName: &scName,
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceName(v1.ResourceStorage): size,
					},
				},
			},
		},
	}
}

func (d *Deployment) createPodTemplateSpec(nfsPort int, rpcPort int) corev1.PodTemplateSpec {

	return v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			// Name:      d.nfsServer.Name,
			// Namespace: d.nfsServer.Namespace,
			Labels: labelsForStatefulSet(d.nfsServer.Name),
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					ImagePullPolicy: "IfNotPresent",
					Name:            "ganesha",
					// Name:            d.nfsServer.Name,
					Image: d.nfsServer.Spec.GetContainerImage(),
					// Args: []string{"nfs", "server", "--ganeshaConfigPath=" + NFSConfigMapPath + "/" + nfsServer.name},
					Ports: []v1.ContainerPort{
						{
							Name:          "nfs-port",
							ContainerPort: int32(nfsPort),
						},
						{
							Name:          "rpc-port",
							ContainerPort: int32(rpcPort),
						},
					},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "nfs-data",
							MountPath: "/export",
						},
					},
					SecurityContext: &v1.SecurityContext{
						Capabilities: &v1.Capabilities{
							Add: []v1.Capability{
								"SYS_ADMIN",
								"DAC_READ_SEARCH",
							},
						},
					},
				},
			},
		},
	}
}

func (d *Deployment) deleteStatefulSet(name string) error {
	return d.deleteObject(d.getStatefulSet(name))
}

func (d *Deployment) getStatefulSet(name string) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: d.nfsServer.Namespace,
			Labels: map[string]string{
				"app": "storageos",
			},
		},
	}
}
