package testutils

import (
	"encoding/base64"

	korifiv1alpha1 "code.cloudfoundry.org/korifi/controllers/api/v1alpha1"
	"code.cloudfoundry.org/korifi/tools"

	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GenerateGUID() string {
	return uuid.NewString()
}

func PrefixedGUID(prefix string) string {
	return prefix + "-" + uuid.NewString()[:8]
}

func BuildCFAppEnvVarsSecret(appGUID, spaceGUID string, envVars map[string]string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: spaceGUID,
			Name:      appGUID + "-env",
		},
		StringData: envVars,
	}
}

func BuildCFBuildObject(cfBuildGUID string, namespace string, cfPackageGUID string, cfAppGUID string) *korifiv1alpha1.CFBuild {
	return &korifiv1alpha1.CFBuild{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfBuildGUID,
			Namespace: namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "CFBuild",
			APIVersion: "korifi.cloudfoundry.org/v1alpha1",
		},
		Spec: korifiv1alpha1.CFBuildSpec{
			PackageRef: corev1.LocalObjectReference{
				Name: cfPackageGUID,
			},
			AppRef: corev1.LocalObjectReference{
				Name: cfAppGUID,
			},
			StagingMemoryMB: 1024,
			StagingDiskMB:   1024,
			Lifecycle: korifiv1alpha1.Lifecycle{
				Type: "buildpack",
				Data: korifiv1alpha1.LifecycleData{
					Buildpacks: nil,
					Stack:      "",
				},
			},
		},
	}
}

func BuildCFBuildDropletStatusObject(dropletProcessTypeMap map[string]string) *korifiv1alpha1.BuildDropletStatus {
	dropletProcessTypes := make([]korifiv1alpha1.ProcessType, 0, len(dropletProcessTypeMap))
	for k, v := range dropletProcessTypeMap {
		dropletProcessTypes = append(dropletProcessTypes, korifiv1alpha1.ProcessType{
			Type:    k,
			Command: v,
		})
	}
	return &korifiv1alpha1.BuildDropletStatus{
		Registry: korifiv1alpha1.Registry{
			Image:            "image/registry/url",
			ImagePullSecrets: []corev1.LocalObjectReference{{Name: "some-image-pull-secret"}},
		},
		Stack:        "cflinuxfs3",
		ProcessTypes: dropletProcessTypes,
	}
}

func BuildDockerRegistrySecret(name, namespace string) *corev1.Secret {
	dockerRegistryUsername := "user"
	dockerRegistryPassword := "password"
	dockerAuth := base64.StdEncoding.EncodeToString([]byte(dockerRegistryUsername + ":" + dockerRegistryPassword))
	dockerConfigJSON := `{"auths":{"https://index.docker.io/v1/":{"username":"` + dockerRegistryUsername + `","password":"` + dockerRegistryPassword + `","auth":"` + dockerAuth + `"}}}`
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Immutable: nil,
		Data:      nil,
		StringData: map[string]string{
			".dockerconfigjson": dockerConfigJSON,
		},
		Type: "kubernetes.io/dockerconfigjson",
	}
}

func BuildServiceAccount(name, namespace, imagePullSecretName string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Secrets:          []corev1.ObjectReference{{Name: imagePullSecretName}},
		ImagePullSecrets: []corev1.LocalObjectReference{{Name: imagePullSecretName}},
	}
}

func BuildCFProcessCRObject(cfProcessGUID, namespace, cfAppGUID, processType, processCommand, processDetectedCommand string) *korifiv1alpha1.CFProcess {
	return &korifiv1alpha1.CFProcess{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfProcessGUID,
			Namespace: namespace,
			Labels: map[string]string{
				korifiv1alpha1.CFAppGUIDLabelKey:     cfAppGUID,
				korifiv1alpha1.CFProcessGUIDLabelKey: cfProcessGUID,
				korifiv1alpha1.CFProcessTypeLabelKey: processType,
			},
		},
		Spec: korifiv1alpha1.CFProcessSpec{
			AppRef:          corev1.LocalObjectReference{Name: cfAppGUID},
			ProcessType:     processType,
			Command:         processCommand,
			DetectedCommand: processDetectedCommand,
			HealthCheck: korifiv1alpha1.HealthCheck{
				Type: "process",
				Data: korifiv1alpha1.HealthCheckData{
					InvocationTimeoutSeconds: 0,
					TimeoutSeconds:           0,
				},
			},
			DesiredInstances: tools.PtrTo(1),
			MemoryMB:         1024,
			DiskQuotaMB:      100,
		},
	}
}
