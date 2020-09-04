/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"

	nodeprofilev1alpha1 "github.com/nolancon/node-profile-discovery/api/v1alpha1"
)

// NodeProfileReconciler reconciles a NodeProfile object
type NodeProfileReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=nodeprofile.intel.com,resources=nodeprofiles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nodeprofile.intel.com,resources=nodeprofiles/status,verbs=get;update;patch

func (r *NodeProfileReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("nodeprofile", req.NamespacedName)
	logger.Info("nodeprofile_controller reconciling object...")

	// List Nodes in cluster in order to examine labels
	nodeList := &corev1.NodeList{}
	err := r.List(ctx, nodeList)
	if err != nil {
		logger.Info("Failed to list Nodes")
		return ctrl.Result{}, err
	}
	if len(nodeList.Items) == 0 {
		logger.Info("No Nodes found")
		return ctrl.Result{}, nil
	}

	// Get NodeProfile object
	nodeProfile := &nodeprofilev1alpha1.NodeProfile{}
	err = r.Get(ctx, req.NamespacedName, nodeProfile)
	if err != nil {
		if errors.IsNotFound(err) {
			// NodeProfile may have been deleted, in which case corresponding profile labels need
			// to be removed from Nodes.
			for _, node := range nodeList.Items {
				profileLabelKey := fmt.Sprintf("%v%v", "profile.node.kubernetes.io/", req.NamespacedName.Name)
				nodeLabels := labels.Set(node.GetObjectMeta().GetLabels())
				if nodeLabels.Has(profileLabelKey) {
					// Remove node profile label from node
					nodeLabelsMap := map[string]string(nodeLabels)
					delete(nodeLabelsMap, profileLabelKey)
					updatedNodeLabels := labels.Set(nodeLabelsMap)
					node.GetObjectMeta().SetLabels(updatedNodeLabels)
					err := r.Update(ctx, &node)
					if err != nil {
						return ctrl.Result{}, err
					}
					logger.Info("Node Profile label removed successfully", "Node", node.GetObjectMeta().GetName())
				}
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	for _, node := range nodeList.Items {
		nodeLabels := labels.Set(node.GetObjectMeta().GetLabels())
		logger.Info("Node and profile (required) labels", "nodeLabels", nodeLabels, "required labels", nodeProfile.Spec.Labels.Required)

		if !labels.AreLabelsInWhiteList(labels.Set(nodeProfile.Spec.Labels.Required), nodeLabels) {
			logger.Info("Node does not fit profile for <required> labels")
			continue
		}

		logger.Info("Node contains all <required> labels. Create 'best-effort' profile label...")

		// Create best-effort label
		profileLabelKey := fmt.Sprintf("%v%v", "profile.node/", nodeProfile.GetObjectMeta().GetName())
		profileLabel := labels.Set{
			profileLabelKey: "best-effort",
		}

		if labels.AreLabelsInWhiteList(labels.Set(nodeProfile.Spec.Labels.Preferred), nodeLabels) {
			logger.Info("Node also contains all <preferred> labels. Update profile label to 'strict'...")
			profileLabel = labels.Set{
				profileLabelKey: "strict",
			}
		}

		// Update node labels with profile label if it does not already exist OR if it has a different value
		if nodeLabels.Get(profileLabelKey) == "" || labels.Conflicts(profileLabel, nodeLabels) {
			logger.Info("Apply profile label to node...")
			updatedLabels := labels.Merge(nodeLabels, profileLabel)
			node.GetObjectMeta().SetLabels(updatedLabels)
			err := r.Update(ctx, &node)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	}
	logger.Info("NodeProfile Reconciled.")
	// Requeue after n seconds (stnadard NFD interval in 60 seconds
	return ctrl.Result{RequeueAfter: time.Second * 5}, nil
}

func (r *NodeProfileReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nodeprofilev1alpha1.NodeProfile{}).
		Complete(r)
}
