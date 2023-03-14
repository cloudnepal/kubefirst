package k8s

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
)

// VerifyArgoCDReadiness waits for critical resources within ArgoCD to be ready
// and only returns once they're all healthy
//
// This helps prevent race conditions and timeouts
func VerifyArgoCDReadiness(kubeconfigPath string, highAvailabilityEnabled bool) (bool, error) {
	// argocd-application-controller StatefulSet
	argoCDStatefulSet, err := ReturnStatefulSetObject(
		kubeconfigPath,
		"app.kubernetes.io/name",
		"argocd-application-controller",
		"argocd",
		120,
	)
	if err != nil {
		return false, errors.New(fmt.Sprintf("Error finding ArgoCD Application Controller StatefulSet: %s", err))
	}
	_, err = WaitForStatefulSetReady(kubeconfigPath, argoCDStatefulSet, 120, false)
	if err != nil {
		return false, errors.New(fmt.Sprintf("Error waiting for ArgoCD Application Controller StatefulSet ready state: %s", err))
	}

	// argocd-server Deployment
	argoCDServerDeployment, err := ReturnDeploymentObject(
		kubeconfigPath,
		"app.kubernetes.io/name",
		"argocd-server",
		"argocd",
		120,
	)
	if err != nil {
		log.Info().Msgf("Error finding ArgoCD server deployment: %s", err)
	}
	_, err = WaitForDeploymentReady(kubeconfigPath, argoCDServerDeployment, 120)
	if err != nil {
		log.Info().Msgf("Error waiting for ArgoCD server deployment ready state: %s", err)
	}

	// Wait for additional ArgoCD Pods to transition to Running
	// This is related to a condition where apps attempt to deploy before
	// repo, redis, or other health checks are passing
	//
	// This can cause future steps to break since the registry app
	// may never apply

	// argocd-repo-server
	argoCDRepoDeployment, err := ReturnDeploymentObject(
		kubeconfigPath,
		"app.kubernetes.io/name",
		"argocd-repo-server",
		"argocd",
		120,
	)
	if err != nil {
		return false, errors.New(fmt.Sprintf("Error finding ArgoCD repo deployment: %s", err))
	}
	_, err = WaitForDeploymentReady(kubeconfigPath, argoCDRepoDeployment, 120)
	if err != nil {
		return false, errors.New(fmt.Sprintf("Error waiting for ArgoCD repo deployment ready state: %s", err))
	}

	// high availability components
	if highAvailabilityEnabled {
		// argocd-redis-ha-haproxy Deployment
		argoCDRedisHAhaproxyDeployment, err := ReturnDeploymentObject(
			kubeconfigPath,
			"app.kubernetes.io/name",
			"argocd-redis-ha-haproxy",
			"argocd",
			120,
		)
		if err != nil {
			return false, errors.New(fmt.Sprintf("Error finding ArgoCD argocd-redis-ha-haproxy Deployment: %s", err))
		}
		_, err = WaitForDeploymentReady(kubeconfigPath, argoCDRedisHAhaproxyDeployment, 120)
		if err != nil {
			return false, errors.New(fmt.Sprintf("Error waiting for ArgoCD argocd-redis-ha-haproxy deployment ready state: %s", err))
		}

		// argocd-redis-ha StatefulSet
		argoCDRedisHAServerStatefulSet, err := ReturnStatefulSetObject(
			kubeconfigPath,
			"app.kubernetes.io/name",
			"argocd-redis-ha",
			"argocd",
			120,
		)
		if err != nil {
			return false, errors.New(fmt.Sprintf("Error finding ArgoCD argocd-redis-ha StatefulSet: %s", err))
		}
		_, err = WaitForStatefulSetReady(kubeconfigPath, argoCDRedisHAServerStatefulSet, 120, false)
		if err != nil {
			return false, errors.New(fmt.Sprintf("Error waiting for ArgoCD argocd-redis-ha StatefulSet ready state: %s", err))
		}
	} else {
		// non-high availability components
		// argocd-redis Deployment
		argoCDRedisDeployment, err := ReturnDeploymentObject(
			kubeconfigPath,
			"app.kubernetes.io/name",
			"argocd-redis",
			"argocd",
			120,
		)
		if err != nil {
			return false, errors.New(fmt.Sprintf("Error finding ArgoCD argocd-redis Deployment: %s", err))
		}
		_, err = WaitForDeploymentReady(kubeconfigPath, argoCDRedisDeployment, 120)
		if err != nil {
			return false, errors.New(fmt.Sprintf("Error waiting for ArgoCD argocd-redis Deployment ready state: %s", err))
		}
	}

	return true, nil
}