/*
Copyright 2021 Absa Group Limited

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Generated by GoLic, for more details see: https://github.com/AbsaOSS/golic
*/
package test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
)

// Basic k8gb deployment test that is verifying that associated ingress is getting created
func TestK8gbIngressAnnotationFailover(t *testing.T) {
	t.Parallel()

	// Path to the Kubernetes resource config we will test
	kubeResourcePath, err := filepath.Abs("../examples/ingress-annotation-failover.yaml")
	require.NoError(t, err)
	brokenResourcePath, err := filepath.Abs("../examples/broken-ingress-annotation.yaml")
	require.NoError(t, err)

	// To ensure we can reuse the resource config on the same cluster to test different scenarios, we setup a unique
	// namespace for the resources for this test.
	// Note that namespaces must be lowercase.
	namespaceName := fmt.Sprintf("k8gb-basic-example-%s", strings.ToLower(random.UniqueId()))

	// Here we choose to use the defaults, which is:
	// - HOME/.kube/config for the kubectl config file
	// - Current context of the kubectl config file
	// - Random namespace
	options := k8s.NewKubectlOptions("", "", namespaceName)

	k8s.CreateNamespace(t, options, namespaceName)

	defer k8s.DeleteNamespace(t, options, namespaceName)

	defer k8s.KubectlDelete(t, options, kubeResourcePath)

	k8s.KubectlApply(t, options, kubeResourcePath)

	ingress := k8s.GetIngress(t, options, "test-gslb-annotation-failover")
	require.Equal(t, ingress.Name, "test-gslb-annotation-failover")
	assertGslbStatus(t, options, "test-gslb-annotation-failover", "notfound.cloud.example.com:NotFound roundrobin.cloud.example.com:NotFound unhealthy.cloud.example.com:NotFound")
	assertGslbSpec(t, options, "test-gslb-annotation-failover", ".spec.strategy.type", "failover")
	assertGslbSpec(t, options, "test-gslb-annotation-failover", ".spec.strategy.primaryGeoTag", "eu")

	t.Run("Broken ingress is not proccessed", func(t *testing.T) {
		k8s.KubectlApply(t, options, brokenResourcePath)
		err := k8s.RunKubectlE(t, options, "get", "gslb", "broken-test-gslb-annotation-failover")
		require.Error(t, err)
	})
}
