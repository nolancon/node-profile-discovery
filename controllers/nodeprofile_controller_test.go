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
	"reflect"
	"testing"

	nodeprofilev1alpha1 "github.com/nolancon/node-profile-discovery/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func createNodeProfileReconcilerObject(nodeProfile *nodeprofilev1alpha1.NodeProfile) (*NodeProfileReconciler, error) {
	// Register operator types with the runtime scheme.
	s := scheme.Scheme

	// Add route Openshift scheme
	if err := nodeprofilev1alpha1.AddToScheme(s); err != nil {
		return nil, err
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{nodeProfile}

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	// Create a ReconcileNode object with the scheme and fake client.
	r := &NodeProfileReconciler{cl, ctrl.Log.WithName("testing"), s}

	return r, nil

}

func TestReconcile(t *testing.T) {
	//TODO: Add more test cases.
	tcases := []struct {
		name                string
		nodeProfile         *nodeprofilev1alpha1.NodeProfile
		nodeList            *corev1.NodeList
		expectedNodeProfile *nodeprofilev1alpha1.NodeProfile
		expectedNodeList    *corev1.NodeList
	}{
		{
			name: "test case 1 - single node, one required label",
			nodeProfile: &nodeprofilev1alpha1.NodeProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "node-profile-1",
					Namespace: "default",
				},
				Spec: nodeprofilev1alpha1.NodeProfileSpec{
					Labels: nodeprofilev1alpha1.NodeLabels{
						Required: map[string]string{
							"label1key": "label1val",
						},
					},
				},
			},
			nodeList: &corev1.NodeList{
				Items: []corev1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "node-1",
							Namespace: "default",
							Labels: map[string]string{
								"label1key": "label1val",
							},
						},
						Spec: corev1.NodeSpec{},
					},
				},
			},
			expectedNodeProfile: &nodeprofilev1alpha1.NodeProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "node-profile-1",
					Namespace: "default",
				},
				Spec: nodeprofilev1alpha1.NodeProfileSpec{},
				Status: nodeprofilev1alpha1.NodeProfileStatus{
					Nodes: map[string]string{
						"node-1": "strict",
					},
				},
			},
			expectedNodeList: &corev1.NodeList{
				Items: []corev1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "node-1",
							Namespace: "default",
							Labels: map[string]string{
								"label1key":                   "label1val",
								"profile.node/node-profile-1": "strict",
							},
						},
						Spec: corev1.NodeSpec{},
					},
				},
			},
		},
		{
			name: "test case 2 - 2 nodes, 1 best-effort, 1 strict",
			nodeProfile: &nodeprofilev1alpha1.NodeProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "node-profile-1",
					Namespace: "default",
				},
				Spec: nodeprofilev1alpha1.NodeProfileSpec{
					Labels: nodeprofilev1alpha1.NodeLabels{
						Required: map[string]string{
							"label1key": "label1val",
						},
						Preferred: map[string]string{
							"label2key": "label2val",
						},
					},
				},
			},
			nodeList: &corev1.NodeList{
				Items: []corev1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "node-1",
							Namespace: "default",
							Labels: map[string]string{
								"label1key": "label1val",
							},
						},
						Spec: corev1.NodeSpec{},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "node-2",
							Namespace: "default",
							Labels: map[string]string{
								"label1key": "label1val",
								"label2key": "label2val",
							},
						},
						Spec: corev1.NodeSpec{},
					},
				},
			},
			expectedNodeProfile: &nodeprofilev1alpha1.NodeProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "node-profile-1",
					Namespace: "default",
				},
				Spec: nodeprofilev1alpha1.NodeProfileSpec{},
				Status: nodeprofilev1alpha1.NodeProfileStatus{
					Nodes: map[string]string{
						"node-1": "best-effort",
						"node-2": "strict",
					},
				},
			},
			expectedNodeList: &corev1.NodeList{
				Items: []corev1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "node-1",
							Namespace: "default",
							Labels: map[string]string{
								"label1key":                   "label1val",
								"profile.node/node-profile-1": "best-effort",
							},
						},
						Spec: corev1.NodeSpec{},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "node-1",
							Namespace: "default",
							Labels: map[string]string{
								"label1key":                   "label1val",
								"label2key":                   "label2val",
								"profile.node/node-profile-1": "strict",
							},
						},
						Spec: corev1.NodeSpec{},
					},
				},
			},
		},
		{
			name: "test case 3 - multiple nodes",
			nodeProfile: &nodeprofilev1alpha1.NodeProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "node-profile-1",
					Namespace: "default",
				},
				Spec: nodeprofilev1alpha1.NodeProfileSpec{
					Labels: nodeprofilev1alpha1.NodeLabels{
						Required: map[string]string{
							"label1key": "label1val",
							"label2key": "label2val",
							"label3key": "label3val",
							"label4key": "label4val",
							"label5key": "label5val",
						},
						Preferred: map[string]string{
							"label6key":  "label6val",
							"label7key":  "label7val",
							"label8key":  "label8val",
							"label9key":  "label9val",
							"label10key": "label10val",
						},
					},
				},
			},
			nodeList: &corev1.NodeList{
				Items: []corev1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "node-1",
							Namespace: "default",
							Labels: map[string]string{
								"label1key": "label1val",
							},
						},
						Spec: corev1.NodeSpec{},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "node-2",
							Namespace: "default",
							// Some required, no preferred
							Labels: map[string]string{
								"label1key": "label1val",
								"label3key": "label3val",
								"label4key": "label4val",
								"label5key": "label5val",
							},
						},
						Spec: corev1.NodeSpec{},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "node-3",
							Namespace: "default",
							// All required, no preferred
							Labels: map[string]string{
								"label1key": "label1val",
								"label2key": "label2val",
								"label3key": "label3val",
								"label4key": "label4val",
								"label5key": "label5val",
							},
						},
						Spec: corev1.NodeSpec{},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "node-4",
							Namespace: "default",
							// All required, some preferred
							Labels: map[string]string{
								"label1key": "label1val",
								"label2key": "label2val",
								"label3key": "label3val",
								"label4key": "label4val",
								"label5key": "label5val",
								"label6key": "label6val",
								"label7key": "label7val",
								"label8key": "label8val",
							},
						},
						Spec: corev1.NodeSpec{},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "node-5",
							Namespace: "default",
							// All required, all preferred
							Labels: map[string]string{
								"label1key":  "label1val",
								"label2key":  "label2val",
								"label3key":  "label3val",
								"label4key":  "label4val",
								"label5key":  "label5val",
								"label6key":  "label6val",
								"label7key":  "label7val",
								"label8key":  "label8val",
								"label9key":  "label9val",
								"label10key": "label10val",
							},
						},
						Spec: corev1.NodeSpec{},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "node-6",
							Namespace: "default",
							// Some required, all preferred
							Labels: map[string]string{
								"label1key":  "label1val",
								"label5key":  "label5val",
								"label6key":  "label6val",
								"label7key":  "label7val",
								"label8key":  "label8val",
								"label9key":  "label9val",
								"label10key": "label10val",
							},
						},
						Spec: corev1.NodeSpec{},
					},
				},
			},
			expectedNodeProfile: &nodeprofilev1alpha1.NodeProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "node-profile-1",
					Namespace: "default",
				},
				Spec: nodeprofilev1alpha1.NodeProfileSpec{},
				Status: nodeprofilev1alpha1.NodeProfileStatus{
					Nodes: map[string]string{
						"node-3": "best-effort",
						"node-4": "best-effort",
						"node-5": "strict",
					},
				},
			},
			expectedNodeList: &corev1.NodeList{
				Items: []corev1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "node-1",
							Namespace: "default",
							Labels: map[string]string{
								"label1key": "label1val",
							},
						},
						Spec: corev1.NodeSpec{},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "node-2",
							Namespace: "default",
							// Some required, no preferred
							Labels: map[string]string{
								"label1key": "label1val",
								"label3key": "label3val",
								"label4key": "label4val",
								"label5key": "label5val",
							},
						},
						Spec: corev1.NodeSpec{},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "node-3",
							Namespace: "default",
							// All required, no preferred
							Labels: map[string]string{
								"label1key":                   "label1val",
								"label2key":                   "label2val",
								"label3key":                   "label3val",
								"label4key":                   "label4val",
								"label5key":                   "label5val",
								"profile.node/node-profile-1": "best-effort",
							},
						},
						Spec: corev1.NodeSpec{},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "node-4",
							Namespace: "default",
							// All required, some preferred
							Labels: map[string]string{
								"label1key":                   "label1val",
								"label2key":                   "label2val",
								"label3key":                   "label3val",
								"label4key":                   "label4val",
								"label5key":                   "label5val",
								"label6key":                   "label6val",
								"label7key":                   "label7val",
								"label8key":                   "label8val",
								"profile.node/node-profile-1": "best-effort",
							},
						},
						Spec: corev1.NodeSpec{},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "node-5",
							Namespace: "default",
							// All required, all preferred
							Labels: map[string]string{
								"label1key":                   "label1val",
								"label2key":                   "label2val",
								"label3key":                   "label3val",
								"label4key":                   "label4val",
								"label5key":                   "label5val",
								"label6key":                   "label6val",
								"label7key":                   "label7val",
								"label8key":                   "label8val",
								"label9key":                   "label9val",
								"label10key":                  "label10val",
								"profile.node/node-profile-1": "strict",
							},
						},
						Spec: corev1.NodeSpec{},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "node-6",
							Namespace: "default",
							// Some required, all preferred
							Labels: map[string]string{
								"label1key":  "label1val",
								"label5key":  "label5val",
								"label6key":  "label6val",
								"label7key":  "label7val",
								"label8key":  "label8val",
								"label9key":  "label9val",
								"label10key": "label10val",
							},
						},
						Spec: corev1.NodeSpec{},
					},
				},
			},
		},
	}

	for _, tc := range tcases {
		// Create a ReconcileNode object with the scheme and fake client.
		r, err := createNodeProfileReconcilerObject(tc.nodeProfile)
		if err != nil {
			t.Fatalf("error creating ReconcileNodeProfile object: (%v)", err)
		}

		nodeProfileName := tc.nodeProfile.GetObjectMeta().GetName()
		nodeProfileNamespace := tc.nodeProfile.GetObjectMeta().GetNamespace()
		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      nodeProfileName,
				Namespace: nodeProfileNamespace,
			},
		}

		for _, node := range tc.nodeList.Items {
			r.Create(context.TODO(), &node)
			if err != nil {
				t.Fatalf("Failed to create test node")
			}
		}
		res, err := r.Reconcile(req)
		if err != nil {
			t.Fatalf("reconcile: (%v)", err)
		}
		//Check the result of reconciliation to make sure it has the desired state.
		if res.Requeue {
			t.Error("reconcile unexpectedly requeued request")
		}

		nodeProfile := &nodeprofilev1alpha1.NodeProfile{}
		err = r.Get(context.TODO(), req.NamespacedName, nodeProfile)
		if err != nil {
			t.Fatalf("Failed to retrieve updated nodestate")
		}

		if !reflect.DeepEqual(tc.expectedNodeProfile.Status.Nodes, nodeProfile.Status.Nodes) {
			t.Errorf("Failed: %v - Expected %v, got %v", tc.name, tc.expectedNodeProfile.Status.Nodes, nodeProfile.Status.Nodes)
		}

		nodeList := &corev1.NodeList{}
		err = r.List(context.TODO(), nodeList)
		if err != nil {
			t.Fatalf("Failed to retrieve updated nodestate")
		}

		node := &corev1.Node{}
		nsm := types.NamespacedName{
			Name:      "node-1",
			Namespace: "default",
		}
		err = r.Get(context.TODO(), nsm, node)
		if err != nil {
			t.Fatalf("Failed to retrieve updated nodestate")
		}

		for i, nodeListItem := range tc.nodeList.Items {
			expLabels := tc.expectedNodeList.Items[i].GetObjectMeta().GetLabels()
			node := &corev1.Node{}
			nodeNamespacedName := types.NamespacedName{
				Name:      nodeListItem.GetObjectMeta().GetName(),
				Namespace: nodeListItem.GetObjectMeta().GetNamespace(),
			}
			err = r.Get(context.TODO(), nodeNamespacedName, node)
			if err != nil {
				t.Fatalf("Failed to retrieve updated node")
			}

			actualLabels := node.GetObjectMeta().GetLabels()
			if !reflect.DeepEqual(expLabels, actualLabels) {
				t.Errorf("Failed: %v - Expected %v, got %v", tc.name, expLabels, actualLabels)
			}
		}

	}
}
