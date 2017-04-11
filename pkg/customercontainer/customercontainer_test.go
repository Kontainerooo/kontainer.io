package customercontainer_test

import (
	"context"
	"os"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/kontainerooo/kontainer.ooo/pkg/customercontainer"
	"github.com/kontainerooo/kontainer.ooo/pkg/kmi"
	"github.com/kontainerooo/kontainer.ooo/pkg/testutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Customercontainer", func() {

	Describe("Create service", func() {
		It("Should create a new service", func() {
			cc, err := customercontainer.NewService(testutils.NewMockDCli(), testutils.NewMockDB())
			Ω(cc).ShouldNot(BeZero())
			Ω(err).ShouldNot(HaveOccurred())
		})
	})

	Describe("Create Container", func() {
		db := testutils.NewMockDB()
		mockKMI := testutils.NewMockKMIClient()
		mockKMIEndpoints := testutils.NewMockKMIEndpoints(log.NewNopLogger(), *mockKMI)
		mockKMI.AddMockKMI(0, kmi.KMI{
			KMDI: kmi.KMDI{
				ID:          1,
				Name:        "node",
				Version:     "",
				Description: "",
				Type:        3,
			},
			Dockerfile:  "FROM FROM node:7-wheezy",
			Context:     "./container-test",
			Commands:    nil,
			Environment: nil,
			Frontend:    nil,
			Imports:     nil,
			Interfaces:  nil,
			Mounts:      nil,
			Variables:   nil,
			Resources: map[string]interface{}{
				"cpus": 1,
				"mem":  500,
				"swap": 500,
			},
		})

		cli := testutils.NewMockDCli()
		cc, _ := customercontainer.NewService(cli, db)
		cc.AddLogger(log.NewNopLogger())
		cc.AddKMIClient(mockKMIEndpoints)

		It("Should create a new container", func() {
			containerName, _, err := cc.CreateContainer(123, 0)

			Ω(err).ShouldNot(HaveOccurred())
			Ω(strings.HasPrefix(containerName, "123")).Should(BeTrue())
		})

		It("Should fail with missing kmi", func() {
			containerName, _, err := cc.CreateContainer(123, 1)

			Ω(err).Should(HaveOccurred())
			Ω(containerName).Should(BeZero())
		})

		It("Should fail creating the docker container", func() {
			cli := testutils.NewMockDCli()
			cc, _ := customercontainer.NewService(cli, db)
			cli.CreateMockImage("testimage")

			cli.SetDockerOffline()
			containerName, _, err := cc.CreateContainer(123, 0)

			Ω(err).Should(HaveOccurred())
			Ω(containerName).Should(BeZero())
		})

		It("Should fail on renaming the container", func() {
			cli := testutils.NewMockDCli()
			cc, _ := customercontainer.NewService(cli, db)
			cli.CreateMockImage("testimage")

			cli.SetIDNotExisting()
			containerName, _, err := cc.CreateContainer(123, 0)

			Ω(err).Should(HaveOccurred())
			Ω(containerName).Should(BeZero())
		})

		It("Should fail decoding seccomp profile", func() {
			cli := testutils.NewMockDCli()
			cc, _ := customercontainer.NewService(cli, db)
			cli.CreateMockImage("testimage")

			// Save seccomp and remove
			tmpSeccomp := customercontainer.SeccompProfile
			customercontainer.SeccompProfile = ``

			containerName, _, err := cc.CreateContainer(123, 0)

			Ω(err).Should(HaveOccurred())
			Ω(containerName).Should(BeZero())

			customercontainer.SeccompProfile = tmpSeccomp
		})
	})

	Describe("Edit Container", func() {
		db := testutils.NewMockDB()
		cli := testutils.NewMockDCli()
		cc, _ := customercontainer.NewService(cli, db)
		It("Should edit container", func() {
			err := cc.EditContainer("123", &customercontainer.ContainerConfig{})

			Ω(err).ShouldNot(HaveOccurred())
		})
	})

	Describe("Remove container", func() {
		db := testutils.NewMockDB()
		mockKMI := testutils.NewMockKMIClient()
		mockKMIEndpoints := testutils.NewMockKMIEndpoints(log.NewNopLogger(), *mockKMI)
		mockKMI.AddMockKMI(0, kmi.KMI{
			KMDI: kmi.KMDI{
				ID:          1,
				Name:        "node",
				Version:     "",
				Description: "",
				Type:        3,
			},
			Dockerfile:  "FROM FROM node:7-wheezy",
			Context:     "./container-test",
			Commands:    nil,
			Environment: nil,
			Frontend:    nil,
			Imports:     nil,
			Interfaces:  nil,
			Mounts:      nil,
			Variables:   nil,
			Resources: map[string]interface{}{
				"cpus": 1,
				"mem":  500,
				"swap": 500,
			},
		})

		cli := testutils.NewMockDCli()
		cc, _ := customercontainer.NewService(cli, db)
		cc.AddLogger(log.NewNopLogger())
		cc.AddKMIClient(mockKMIEndpoints)

		_, containerID, _ := cc.CreateContainer(123, 0)
		It("Should remove container", func() {
			err := cc.RemoveContainer(containerID)

			Ω(err).ShouldNot(HaveOccurred())
		})

		It("Should fail when container does not exist", func() {
			err := cc.RemoveContainer(containerID)

			Ω(err).Should(HaveOccurred())
		})
	})

	Describe("Get instances", func() {
		db := testutils.NewMockDB()
		mockKMI := testutils.NewMockKMIClient()
		mockKMIEndpoints := testutils.NewMockKMIEndpoints(log.NewNopLogger(), *mockKMI)
		mockKMI.AddMockKMI(0, kmi.KMI{
			KMDI: kmi.KMDI{
				ID:          1,
				Name:        "node",
				Version:     "",
				Description: "",
				Type:        3,
			},
			Dockerfile:  "FROM FROM node:7-wheezy",
			Context:     "./container-test",
			Commands:    nil,
			Environment: nil,
			Frontend:    nil,
			Imports:     nil,
			Interfaces:  nil,
			Mounts:      nil,
			Variables:   nil,
			Resources: map[string]interface{}{
				"cpus": 1,
				"mem":  500,
				"swap": 500,
			},
		})

		cli := testutils.NewMockDCli()
		cc, _ := customercontainer.NewService(cli, db)
		cc.AddLogger(log.NewNopLogger())
		cc.AddKMIClient(mockKMIEndpoints)

		It("Should return instances", func() {
			_, containerID, err := cc.CreateContainer(123, 0)

			instances := cc.Instances(123)

			Ω(err).ShouldNot(HaveOccurred())
			Ω(instances).ShouldNot(BeEmpty())
			Ω(instances[0]).Should(Equal(containerID))

			cc.RemoveContainer(containerID)
		})

		It("Should return no instances when none exist", func() {
			instances := cc.Instances(123)
			Ω(instances).Should(BeEmpty())
		})
	})

	Describe("Create image", func() {
		db := testutils.NewMockDB()
		cli := testutils.NewMockDCli()
		cc, _ := customercontainer.NewService(cli, db)
		cc.AddLogger(log.NewNopLogger())

		os.Mkdir("container-test", 0777)
		os.Create("container-test/.dockerignore")

		mockKMI := testutils.NewMockKMIClient()

		mockKMIEndpoints := testutils.NewMockKMIEndpoints(log.NewNopLogger(), *mockKMI)

		AfterSuite(func() {
			os.RemoveAll("container-test")
		})

		It("Should error when there is no KMI client", func() {
			_, _, err := cc.CreateContainer(123, 0)
			Ω(err).Should(HaveOccurred())

			cc.AddKMIClient(mockKMIEndpoints)
		})

		It("Should create an image", func() {

			mockKMI.AddMockKMI(0, kmi.KMI{
				KMDI: kmi.KMDI{
					ID:          1,
					Name:        "node",
					Version:     "",
					Description: "",
					Type:        3,
				},
				Dockerfile:  "FROM FROM node:7-wheezy",
				Context:     "./container-test",
				Commands:    nil,
				Environment: nil,
				Frontend:    nil,
				Imports:     nil,
				Interfaces:  nil,
				Mounts:      nil,
				Variables:   nil,
				Resources: map[string]interface{}{
					"cpus": 1,
					"mem":  500,
					"swap": 500,
				},
			})

			id, _, err := cc.CreateContainer(123, 0)

			Ω(err).ShouldNot(HaveOccurred())
			Ω(id).ShouldNot(BeNil())
		})

		It("Should fail on invalid environment", func() {
			mockKMI.AddMockKMI(0, kmi.KMI{
				KMDI: kmi.KMDI{
					ID:          1,
					Name:        "node",
					Version:     "",
					Description: "",
					Type:        3,
				},
				Dockerfile: "FROM FROM node:7-wheezy",
				Context:    "./container-test",
				Commands:   nil,
				Environment: map[string]interface{}{
					"in valid": "val",
				},
				Frontend:   nil,
				Imports:    nil,
				Interfaces: nil,
				Mounts:     nil,
				Variables:  nil,
				Resources: map[string]interface{}{
					"cpus": 1,
					"mem":  500,
					"swap": 500,
				},
			})

			_, _, err := cc.CreateContainer(123, 0)
			Ω(err).Should(HaveOccurred())
		})
	})

	Describe("Endpoints and Transport", func() {
		db := testutils.NewMockDB()
		mockKMI := testutils.NewMockKMIClient()
		mockKMIEndpoints := testutils.NewMockKMIEndpoints(log.NewNopLogger(), *mockKMI)
		mockKMI.AddMockKMI(0, kmi.KMI{
			KMDI: kmi.KMDI{
				ID:          1,
				Name:        "node",
				Version:     "",
				Description: "",
				Type:        3,
			},
			Dockerfile:  "FROM FROM node:7-wheezy",
			Context:     "./container-test",
			Commands:    nil,
			Environment: nil,
			Frontend:    nil,
			Imports:     nil,
			Interfaces:  nil,
			Mounts:      nil,
			Variables:   nil,
			Resources: map[string]interface{}{
				"cpus": 1,
				"mem":  500,
				"swap": 500,
			},
		})

		cli := testutils.NewMockDCli()
		cc, _ := customercontainer.NewService(cli, db)
		cc.AddLogger(log.NewNopLogger())
		cc.AddKMIClient(mockKMIEndpoints)

		es := &customercontainer.Endpoints{}
		ctx := context.Background()
		gID := ""

		It("Should create valid Endpoints", func() {
			es.CreateContainerEndpoint = customercontainer.MakeCreateContainerEndpoint(cc)
			es.EditContainerEndpoint = customercontainer.MakeEditContainerEndpoint(cc)
			es.InstancesEndpoint = customercontainer.MakeInstancesEndpoint(cc)
			es.RemoveContainerEndpoint = customercontainer.MakeRemoveContainerEndpoint(cc)
		})

		Context("CreateContainerEndpoint", func() {
			It("Should work with CreateContainer request and response", func() {
				res, err := es.CreateContainerEndpoint(ctx, customercontainer.CreateContainerRequest{
					RefID: 123,
					KMIID: 0,
				})

				gID = res.(customercontainer.CreateContainerResponse).ID

				Ω(err).ShouldNot(HaveOccurred())
				Ω(res.(customercontainer.CreateContainerResponse).Error).ShouldNot(HaveOccurred())
				Ω(res.(customercontainer.CreateContainerResponse).ID).ShouldNot(BeEmpty())
				Ω(res.(customercontainer.CreateContainerResponse).Name).ShouldNot(BeEmpty())
			})
		})

		Context("EditContainerEndpoint", func() {
			It("Should work with EditContainer request and response", func() {
				cfg := customercontainer.ContainerConfig{
					ImageName: "testimage",
				}
				res, err := es.EditContainerEndpoint(ctx, customercontainer.EditContainerRequest{
					ID:  gID,
					Cfg: &cfg,
				})

				Ω(err).ShouldNot(HaveOccurred())
				Ω(res.(customercontainer.EditContainerResponse).Error).ShouldNot(HaveOccurred())
			})
		})

		Context("InstancesEndpoint", func() {
			It("Should work with Instances request and response", func() {
				res, err := es.InstancesEndpoint(ctx, customercontainer.InstancesRequest{
					RefID: 123,
				})

				Ω(err).ShouldNot(HaveOccurred())
				Ω(res.(customercontainer.InstancesResponse).Instances).ShouldNot(BeNil())
			})
		})

		Context("RemoveContainerEndpoint", func() {
			It("Should work with RemoveContainer request and response", func() {

				res, err := es.RemoveContainerEndpoint(ctx, customercontainer.RemoveContainerRequest{
					ID: gID,
				})

				Ω(err).ShouldNot(HaveOccurred())
				Ω(res.(customercontainer.RemoveContainerResponse).Error).ShouldNot(HaveOccurred())
			})
		})
	})
})
