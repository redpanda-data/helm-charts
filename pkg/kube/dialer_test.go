package kube

import (
	"context"
	"crypto/tls"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/k3s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestDialer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := k3s.Run(ctx, "rancher/k3s:v1.27.1-k3s1")
	require.NoError(t, err)
	t.Cleanup(func() {
		container.Terminate(context.Background())
	})

	config, err := container.GetKubeConfig(ctx)
	require.NoError(t, err)
	restcfg, err := clientcmd.RESTConfigFromKubeConfig(config)
	require.NoError(t, err)
	s := runtime.NewScheme()
	require.NoError(t, clientgoscheme.AddToScheme(s))
	client, err := client.New(restcfg, client.Options{Scheme: s})
	require.NoError(t, err)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "name",
			Namespace: "default",
			Labels: map[string]string{
				"service": "label",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Image: "caddy",
					Name:  "caddy",
					Command: []string{
						"caddy",
					},
					Args: []string{
						"file-server",
						"--domain",
						// use localhost so we don't reach out to an
						// ACME server
						"localhost",
					},
					Ports: []corev1.ContainerPort{{
						Name:          "http",
						ContainerPort: 80,
					}, {
						Name:          "https",
						ContainerPort: 443,
					}},
				},
			},
		},
	}
	err = client.Create(ctx, pod)
	require.NoError(t, err)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "service",
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"service": "label",
			},
			Ports: []corev1.ServicePort{{
				Name: "http",
				Port: 8080,
			}},
		},
	}
	err = client.Create(ctx, service)
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		var ready corev1.Pod
		err := client.Get(ctx, types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, &ready)
		if err != nil {
			return false
		}

		return ready.Status.Phase == corev1.PodRunning
	}, 30*time.Second, 10*time.Millisecond)

	dialer := NewPodDialer(restcfg)
	// Set the `ServerName` to match what Caddy generates for the localhost domain,
	// otherwise it fails due to an SNI mismatch.
	tlsConfig := &tls.Config{InsecureSkipVerify: true, ServerName: "localhost"}
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
			DialContext:     dialer.DialContext,
		},
	}

	for _, host := range []string{
		"http://name.service.default.svc.cluster.local",
		"http://name.service.default.svc",
		"http://name.service.default",
		"http://name.service",
		// https
		"https://name.service.default.svc.cluster.local",
		"https://name.service.default.svc",
		"https://name.service.default",
		"https://name.service",
	} {
		t.Run(host, func(t *testing.T) {
			_, err = httpClient.Get(host)
			require.NoError(t, err)
		})
	}
}
