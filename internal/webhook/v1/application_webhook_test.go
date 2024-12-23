/*
Copyright 2024.

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

package v1

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "github.com/lixd/i-operator/api/v1"
	// TODO (user): Add any additional imports if needed
)

var _ = Describe("Application Webhook", func() {
	var (
		obj       *corev1.Application
		oldObj    *corev1.Application
		validator ApplicationCustomValidator
		defaulter ApplicationCustomDefaulter
	)

	BeforeEach(func() {
		obj = &corev1.Application{}
		oldObj = &corev1.Application{}
		validator = ApplicationCustomValidator{}
		Expect(validator).NotTo(BeNil(), "Expected validator to be initialized")
		defaulter = ApplicationCustomDefaulter{}
		Expect(defaulter).NotTo(BeNil(), "Expected defaulter to be initialized")
		Expect(oldObj).NotTo(BeNil(), "Expected oldObj to be initialized")
		Expect(obj).NotTo(BeNil(), "Expected obj to be initialized")
		// TODO (user): Add any setup logic common to all tests
	})

	AfterEach(func() {
		// TODO (user): Add any teardown logic common to all tests
	})

	Context("When creating Application under Defaulting Webhook", func() {
		// TODO (user): Add logic for defaulting webhooks
		// Example:
		// It("Should apply defaults when a required field is empty", func() {
		//     By("simulating a scenario where defaults should be applied")
		//     obj.SomeFieldWithDefault = ""
		//     By("calling the Default method to apply defaults")
		//     defaulter.Default(ctx, obj)
		//     By("checking that the default values are set")
		//     Expect(obj.SomeFieldWithDefault).To(Equal("default_value"))
		// })
	})

	Context("When creating or updating Application under Validating Webhook", func() {
		// TODO (user): Add logic for validating webhooks
		// Example:
		// It("Should deny creation if a required field is missing", func() {
		//     By("simulating an invalid creation scenario")
		//     obj.SomeRequiredField = ""
		//     Expect(validator.ValidateCreate(ctx, obj)).Error().To(HaveOccurred())
		// })
		//
		// It("Should admit creation if all required fields are present", func() {
		//     By("simulating an invalid creation scenario")
		//     obj.SomeRequiredField = "valid_value"
		//     Expect(validator.ValidateCreate(ctx, obj)).To(BeNil())
		// })
		//
		// It("Should validate updates correctly", func() {
		//     By("simulating a valid update scenario")
		//     oldObj.SomeRequiredField = "updated_value"
		//     obj.SomeRequiredField = "updated_value"
		//     Expect(validator.ValidateUpdate(ctx, oldObj, obj)).To(BeNil())
		// })
	})

})

// TestIsValidImageName 使用表格驱动测试验证镜像名称是否有效
func TestIsValidImageName(t *testing.T) {
	tests := []struct {
		imageName string
		expected  bool
	}{
		{"myrepo/myimage:latest", true},      // 合法镜像
		{"myimage:latest", true},             // 合法镜像
		{"myrepo/my-image:v1.0", true},       // 合法镜像
		{"my_image:v1.0", true},              // 合法镜像
		{"myimage@", false},                  // 不合法镜像，@不允许
		{"myimage", true},                    // 合法镜像，未指定标签
		{"invalid_image:with@symbol", false}, // 不合法镜像，标签中有非法字符@
		{"myrepo/myimage:version_1", true},   // 合法镜像，标签可以包含下划线
		{"my-repo/my_image:version-2", true}, // 合法镜像，标签可以包含短横线
	}

	for _, test := range tests {
		t.Run(test.imageName, func(t *testing.T) {
			result := isValidImageName(test.imageName)
			if result != test.expected {
				t.Errorf("For image '%s', expected %v but got %v", test.imageName, test.expected, result)
			}
		})
	}
}
