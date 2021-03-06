// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package general

import (
	"context"
	"fmt"

	"github.com/gardener/gardener/extensions/pkg/controller/healthcheck"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/gardener/gardener/pkg/utils/kubernetes/health"

	resourcesv1alpha1 "github.com/gardener/gardener-resource-manager/pkg/apis/resources/v1alpha1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ManagedResourceHealthChecker contains all the information for the ManagedResource HealthCheck
type ManagedResourceHealthChecker struct {
	logger              logr.Logger
	seedClient          client.Client
	managedResourceName string
}

// CheckManagedResource is a healthCheck function to check ManagedResources
func CheckManagedResource(managedResourceName string) healthcheck.HealthCheck {
	return &ManagedResourceHealthChecker{
		managedResourceName: managedResourceName,
	}
}

// InjectSeedClient injects the seed client
func (healthChecker *ManagedResourceHealthChecker) InjectSeedClient(seedClient client.Client) {
	healthChecker.seedClient = seedClient
}

// SetLoggerSuffix injects the logger
func (healthChecker *ManagedResourceHealthChecker) SetLoggerSuffix(provider, extension string) {
	healthChecker.logger = log.Log.WithName(fmt.Sprintf("%s-%s-healthcheck-managed-resource", provider, extension))
}

// DeepCopy clones the healthCheck struct by making a copy and returning the pointer to that new copy
func (healthChecker *ManagedResourceHealthChecker) DeepCopy() healthcheck.HealthCheck {
	copy := *healthChecker
	return &copy
}

// Check executes the health check
func (healthChecker *ManagedResourceHealthChecker) Check(ctx context.Context, request types.NamespacedName) (*healthcheck.SingleCheckResult, error) {
	mcmDeployment := &resourcesv1alpha1.ManagedResource{}

	if err := healthChecker.seedClient.Get(ctx, client.ObjectKey{Namespace: request.Namespace, Name: healthChecker.managedResourceName}, mcmDeployment); err != nil {
		err := fmt.Errorf("check Managed Resource failed. Unable to retrieve managed resource '%s' in namespace '%s': %v", healthChecker.managedResourceName, request.Namespace, err)
		healthChecker.logger.Error(err, "Health check failed")
		return nil, err
	}
	if isHealthy, reason, err := managedResourceIsHealthy(mcmDeployment); !isHealthy {
		healthChecker.logger.Error(err, "Health check failed")
		return &healthcheck.SingleCheckResult{
			Status: gardencorev1beta1.ConditionFalse,
			Detail: err.Error(),
			Reason: *reason,
		}, nil
	}

	return &healthcheck.SingleCheckResult{
		Status: gardencorev1beta1.ConditionTrue,
	}, nil
}

func managedResourceIsHealthy(managedResource *resourcesv1alpha1.ManagedResource) (bool, *string, error) {
	if err := health.CheckManagedResource(managedResource); err != nil {
		reason := "ManagedResourceUnhealthy"
		err := fmt.Errorf("managed resource %s in namespace %s is unhealthy: %v", managedResource.Name, managedResource.Namespace, err)
		return false, &reason, err
	}
	return true, nil, nil
}
