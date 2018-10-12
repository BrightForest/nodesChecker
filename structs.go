package main

import (
	"time"
)

type Configuration struct {
	KubernetesApiAddr string
	ClusterAddr string
	AuthToken string
	ClusterApiConnTimeout int
}

type PodList struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Metadata   struct {
		SelfLink        string `json:"selfLink"`
		ResourceVersion string `json:"resourceVersion"`
	} `json:"metadata"`
	Items []struct {
		Metadata struct {
			Name              string    `json:"name"`
			GenerateName      string    `json:"generateName"`
			Namespace         string    `json:"namespace"`
			SelfLink          string    `json:"selfLink"`
			UID               string    `json:"uid"`
			ResourceVersion   string    `json:"resourceVersion"`
			CreationTimestamp time.Time `json:"creationTimestamp"`
			Labels            struct {
				App              string `json:"app"`
				Deployment       string `json:"deployment"`
				Deploymentconfig string `json:"deploymentconfig"`
			} `json:"labels"`
			Annotations struct {
				KubernetesIoCreatedBy                    string `json:"kubernetes.io/created-by"`
				KubernetesIoLimitRanger                  string `json:"kubernetes.io/limit-ranger"`
				OpenshiftIoDeploymentConfigLatestVersion string `json:"openshift.io/deployment-config.latest-version"`
				OpenshiftIoDeploymentConfigName          string `json:"openshift.io/deployment-config.name"`
				OpenshiftIoDeploymentName                string `json:"openshift.io/deployment.name"`
				OpenshiftIoGeneratedBy                   string `json:"openshift.io/generated-by"`
				OpenshiftIoScc                           string `json:"openshift.io/scc"`
			} `json:"annotations"`
			OwnerReferences []struct {
				APIVersion         string `json:"apiVersion"`
				Kind               string `json:"kind"`
				Name               string `json:"name"`
				UID                string `json:"uid"`
				Controller         bool   `json:"controller"`
				BlockOwnerDeletion bool   `json:"blockOwnerDeletion"`
			} `json:"ownerReferences"`
		} `json:"metadata"`
		Spec struct {
			Volumes []struct {
				Name   string `json:"name"`
				Secret struct {
					SecretName  string `json:"secretName"`
					DefaultMode int    `json:"defaultMode"`
				} `json:"secret"`
			} `json:"volumes"`
			Containers []struct {
				Name  string `json:"name"`
				Image string `json:"image"`
				Ports []struct {
					ContainerPort int    `json:"containerPort"`
					Protocol      string `json:"protocol"`
				} `json:"ports"`
				Env []struct {
					Name  string `json:"name"`
					Value string `json:"value"`
				} `json:"env"`
				Resources struct {
					Limits struct {
						CPU    string `json:"cpu"`
						Memory string `json:"memory"`
					} `json:"limits"`
					Requests struct {
						CPU    string `json:"cpu"`
						Memory string `json:"memory"`
					} `json:"requests"`
				} `json:"resources"`
				VolumeMounts []struct {
					Name      string `json:"name"`
					ReadOnly  bool   `json:"readOnly"`
					MountPath string `json:"mountPath"`
				} `json:"volumeMounts"`
				TerminationMessagePath   string `json:"terminationMessagePath"`
				TerminationMessagePolicy string `json:"terminationMessagePolicy"`
				ImagePullPolicy          string `json:"imagePullPolicy"`
				SecurityContext          struct {
					Capabilities struct {
						Drop []string `json:"drop"`
					} `json:"capabilities"`
					Privileged     bool `json:"privileged"`
					SeLinuxOptions struct {
						Level string `json:"level"`
					} `json:"seLinuxOptions"`
				} `json:"securityContext"`
			} `json:"containers"`
			RestartPolicy                 string `json:"restartPolicy"`
			TerminationGracePeriodSeconds int    `json:"terminationGracePeriodSeconds"`
			DNSPolicy                     string `json:"dnsPolicy"`
			NodeSelector                  struct {
				Default string `json:"default"`
			} `json:"nodeSelector"`
			ServiceAccountName string `json:"serviceAccountName"`
			ServiceAccount     string `json:"serviceAccount"`
			NodeName           string `json:"nodeName"`
			SecurityContext    struct {
				SeLinuxOptions struct {
					Level string `json:"level"`
				} `json:"seLinuxOptions"`
				FsGroup int `json:"fsGroup"`
			} `json:"securityContext"`
			ImagePullSecrets []struct {
				Name string `json:"name"`
			} `json:"imagePullSecrets"`
			SchedulerName string `json:"schedulerName"`
		} `json:"spec"`
		Status struct {
			Phase      string `json:"phase"`
			Conditions []struct {
				Type               string      `json:"type"`
				Status             string      `json:"status"`
				LastProbeTime      interface{} `json:"lastProbeTime"`
				LastTransitionTime time.Time   `json:"lastTransitionTime"`
			} `json:"conditions"`
			HostIP            string    `json:"hostIP"`
			PodIP             string    `json:"podIP"`
			StartTime         time.Time `json:"startTime"`
			ContainerStatuses []struct {
				Name  string `json:"name"`
				State struct {
					Running struct {
						StartedAt time.Time `json:"startedAt"`
					} `json:"running"`
				} `json:"state"`
				LastState struct {
				} `json:"lastState"`
				Ready        bool   `json:"ready"`
				RestartCount int    `json:"restartCount"`
				Image        string `json:"image"`
				ImageID      string `json:"imageID"`
				ContainerID  string `json:"containerID"`
			} `json:"containerStatuses"`
			QosClass string `json:"qosClass"`
		} `json:"status"`
	} `json:"items"`
}


type NodeList struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Metadata   struct {
		SelfLink        string `json:"selfLink"`
		ResourceVersion string `json:"resourceVersion"`
	} `json:"metadata"`
	Items []struct {
		Metadata struct {
			Name              string    `json:"name"`
			SelfLink          string    `json:"selfLink"`
			UID               string    `json:"uid"`
			ResourceVersion   string    `json:"resourceVersion"`
			CreationTimestamp time.Time `json:"creationTimestamp"`
			Labels            struct {
				BetaKubernetesIoArch string `json:"beta.kubernetes.io/arch"`
				BetaKubernetesIoOs   string `json:"beta.kubernetes.io/os"`
				Default              string `json:"default"`
				Eip                  string `json:"eip"`
				Gocd                 string `json:"gocd"`
				KubernetesIoHostname string `json:"kubernetes.io/hostname"`
				LoggingEsNode        string `json:"logging-es-node"`
				LoggingInfraFluentd  string `json:"logging-infra-fluentd"`
				Nginx                string `json:"nginx"`
				Region               string `json:"region"`
				Zone                 string `json:"zone"`
			} `json:"labels"`
			Annotations struct {
				VolumesKubernetesIoControllerManagedAttachDetach string `json:"volumes.kubernetes.io/controller-managed-attach-detach"`
			} `json:"annotations"`
		} `json:"metadata"`
		Spec struct {
			ExternalID string `json:"externalID"`
		} `json:"spec"`
		Status struct {
			Capacity struct {
				CPU    string `json:"cpu"`
				Memory string `json:"memory"`
				Pods   string `json:"pods"`
			} `json:"capacity"`
			Allocatable struct {
				CPU    string `json:"cpu"`
				Memory string `json:"memory"`
				Pods   string `json:"pods"`
			} `json:"allocatable"`
			Conditions []struct {
				Type               string    `json:"type"`
				Status             string    `json:"status"`
				LastHeartbeatTime  time.Time `json:"lastHeartbeatTime"`
				LastTransitionTime time.Time `json:"lastTransitionTime"`
				Reason             string    `json:"reason"`
				Message            string    `json:"message"`
			} `json:"conditions"`
			Addresses []struct {
				Type    string `json:"type"`
				Address string `json:"address"`
			} `json:"addresses"`
			DaemonEndpoints struct {
				KubeletEndpoint struct {
					Port int `json:"Port"`
				} `json:"kubeletEndpoint"`
			} `json:"daemonEndpoints"`
			NodeInfo struct {
				MachineID               string `json:"machineID"`
				SystemUUID              string `json:"systemUUID"`
				BootID                  string `json:"bootID"`
				KernelVersion           string `json:"kernelVersion"`
				OsImage                 string `json:"osImage"`
				ContainerRuntimeVersion string `json:"containerRuntimeVersion"`
				KubeletVersion          string `json:"kubeletVersion"`
				KubeProxyVersion        string `json:"kubeProxyVersion"`
				OperatingSystem         string `json:"operatingSystem"`
				Architecture            string `json:"architecture"`
			} `json:"nodeInfo"`
			Images []struct {
				Names     []string `json:"names"`
				SizeBytes int      `json:"sizeBytes"`
			} `json:"images"`
		} `json:"status"`
	} `json:"items"`
}

