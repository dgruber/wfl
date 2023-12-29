package remote_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/dgruber/wfl/pkg/context/remote"
)

var _ = Describe("Remote", func() {

	Context("Remote Context Creation", func() {

		It("should fail when server is not set", func() {
			ctx := NewRemoteContextByCfg(Config{})
			Expect(ctx).NotTo(BeNil())
			Expect(ctx.CtxCreationErr).NotTo(BeNil())
		})

		It("should create a context without basic auth", func() {
			ctx := NewRemoteContextByCfg(Config{
				Server: "http://localhost:8080",
			})
			Expect(ctx).NotTo(BeNil())
			Expect(ctx.CtxCreationErr).To(BeNil())
		})

		It("should create a context with basic auth", func() {
			ctx := NewRemoteContextByCfg(Config{
				Server: "http://localhost:8080",
				BasicAuth: &BasicAuthConfig{
					User:     "user",
					Password: "password",
				},
			})
			Expect(ctx).NotTo(BeNil())
			Expect(ctx.CtxCreationErr).To(BeNil())
		})

	})

})
