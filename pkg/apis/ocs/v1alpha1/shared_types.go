package v1alpha1

import (
	"context"

	oappsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	oimagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	imagev1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OpenShiftObject interface {
	metav1.Object
	runtime.Object
}

type PlatformService interface {
	Create(ctx context.Context, obj runtime.Object) error
	Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error
	List(ctx context.Context, opts *client.ListOptions, list runtime.Object) error
	Update(ctx context.Context, obj runtime.Object) error
	GetCached(ctx context.Context, key client.ObjectKey, obj runtime.Object) error
	ImageStreamTags(namespace string) imagev1.ImageStreamTagInterface
	GetScheme() *runtime.Scheme
	IsMockService() bool
}

type Environment struct {
	Primary []CustomObject `json:"others,omitempty"`
}

type CustomObject struct {
	Omit              bool                       `json:"omit,omitempty"`
	ServiceAccounts   []corev1.ServiceAccount    `json:"serviceAccounts,omitempty"`
	Roles             []rbacv1.Role              `json:"roles,omitempty"`
	RoleBindings      []rbacv1.RoleBinding       `json:"roleBindings,omitempty"`
	DeploymentConfigs []oappsv1.DeploymentConfig `json:"deploymentConfigs,omitempty"`
	StatefulSets      []appsv1.StatefulSet       `json:"statefulSets,omitempty"`
	BuildConfigs      []buildv1.BuildConfig      `json:"buildConfigs,omitempty"`
	ImageStreams      []oimagev1.ImageStream     `json:"imageStreams,omitempty"`
	Services          []corev1.Service           `json:"services,omitempty"`
	Routes            []routev1.Route            `json:"routes,omitempty"`
}

type EnvTemplate struct {
	*CommonConfig `json:",inline"`
	Console       ConsoleTemplate     `json:"console,omitempty"`
	Servers       []ServerTemplate    `json:"servers,omitempty"`
	SmartRouter   SmartRouterTemplate `json:"smartRouter,omitempty"`
	Auth          AuthTemplate        `json:"auth,omitempty"`
	Constants     TemplateConstants   `json:"constants,omitempty"`
}

// CommonConfig variables used in the templates
type CommonConfig struct {
	ApplicationName    string `json:"applicationName,omitempty"`
	Version            string `json:"version,omitempty"`
	ImageTag           string `json:"imageTag,omitempty"`
}