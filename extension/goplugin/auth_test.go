// Copyright (C) 2017 NTT Innovation Institute, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package goplugin_test

import (
	"github.com/cloudwan/gohan/extension/goext"
	"github.com/cloudwan/gohan/extension/goplugin"
	"github.com/cloudwan/gohan/schema"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	abstractSchemaPath = "../../tests/test_abstract_schema.yaml"
	schemaPath         = "../../tests/test_schema.yaml"
	adminTenantID      = "12345678aaaaaaaaaaaa123456789012"
	demoTenantID       = "12345678bbbbbbbbbbbb123456789012"
)

var _ = Describe("Auth", func() {

	var (
		manager         *schema.Manager
		adminAuth       schema.Authorization
		adminOnDemoAuth schema.Authorization
		memberAuth      schema.Authorization
		env             goext.IEnvironment
	)

	BeforeEach(func() {
		manager = schema.GetManager()

		Expect(manager.LoadSchemaFromFile(abstractSchemaPath)).To(Succeed())
		Expect(manager.LoadSchemaFromFile(schemaPath)).To(Succeed())

		adminAuth = schema.NewAuthorization(adminTenantID, "admin", "fake_token", []string{"admin"}, nil)
		adminOnDemoAuth = schema.NewAuthorization(adminTenantID, "demo", "fake_token", []string{"admin"}, nil)
		memberAuth = schema.NewAuthorization(demoTenantID, "demo", "fake_token", []string{"Member"}, nil)

		env = goplugin.NewEnvironment("test", nil, nil)
	})

	AfterEach(func() {
		schema.ClearManager()
	})

	setup := func(auth schema.Authorization) goext.Context {
		policy, role := manager.PolicyValidate("create", "/v2.0/networks", auth)
		Expect(policy).NotTo(BeNil())

		context := goext.MakeContext()
		context["policy"] = policy
		context["role"] = role
		context["tenant_id"] = auth.TenantID()
		context["tenant_name"] = auth.TenantName()
		context["auth_token"] = auth.AuthToken()
		context["catalog"] = auth.Catalog()
		context["auth"] = auth

		return context
	}

	It("Return true for admin context", func() {
		context := setup(adminAuth)
		Expect(env.Auth().IsAdminContext(context)).To(BeTrue())
	})

	It("Return false for member context", func() {
		context := setup(memberAuth)
		Expect(env.Auth().IsAdminContext(context)).To(BeFalse())
	})

	It("Return false for admin user logged in as other tenant", func() {
		context := setup(adminOnDemoAuth)
		Expect(env.Auth().IsAdminContext(context)).To(BeFalse())
	})
})
