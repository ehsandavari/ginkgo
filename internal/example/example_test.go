package example_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"

	. "github.com/onsi/ginkgo/internal/example"

	"github.com/onsi/ginkgo/internal/codelocation"
	"github.com/onsi/ginkgo/internal/containernode"
	Failer "github.com/onsi/ginkgo/internal/failer"
	"github.com/onsi/ginkgo/internal/leafnodes"
	"github.com/onsi/ginkgo/internal/types"
	"github.com/onsi/ginkgo/types"
)

var noneFlag = internaltypes.FlagTypeNone
var focusedFlag = internaltypes.FlagTypeFocused
var pendingFlag = internaltypes.FlagTypePending

var _ = Describe("Example", func() {
	var (
		failer       *Failer.Failer
		codeLocation types.CodeLocation
		nodesThatRan []string
		example      *Example
	)

	newBody := func(text string, fail bool) func() {
		return func() {
			nodesThatRan = append(nodesThatRan, text)
			if fail {
				failer.Fail(text, codeLocation)
			}
		}
	}

	newIt := func(text string, flag internaltypes.FlagType, fail bool) *leafnodes.ItNode {
		return leafnodes.NewItNode(text, newBody(text, fail), flag, codeLocation, 0, failer, 0)
	}

	newItWithBody := func(text string, body interface{}) *leafnodes.ItNode {
		return leafnodes.NewItNode(text, body, noneFlag, codeLocation, 0, failer, 0)
	}

	newMeasure := func(text string, flag internaltypes.FlagType, fail bool, samples int) *leafnodes.MeasureNode {
		return leafnodes.NewMeasureNode(text, func(Benchmarker) {
			nodesThatRan = append(nodesThatRan, text)
			if fail {
				failer.Fail(text, codeLocation)
			}
		}, flag, codeLocation, samples, failer, 0)
	}

	newBef := func(text string, fail bool) internaltypes.BasicNode {
		return leafnodes.NewBeforeEachNode(newBody(text, fail), codeLocation, 0, failer, 0)
	}

	newAft := func(text string, fail bool) internaltypes.BasicNode {
		return leafnodes.NewAfterEachNode(newBody(text, fail), codeLocation, 0, failer, 0)
	}

	newJusBef := func(text string, fail bool) internaltypes.BasicNode {
		return leafnodes.NewJustBeforeEachNode(newBody(text, fail), codeLocation, 0, failer, 0)
	}

	newContainer := func(text string, flag internaltypes.FlagType, setupNodes ...internaltypes.BasicNode) *containernode.ContainerNode {
		c := containernode.New(text, flag, codeLocation)
		for _, node := range setupNodes {
			switch node.Type() {
			case types.ExampleComponentTypeBeforeEach:
				c.PushBeforeEachNode(node)
			case types.ExampleComponentTypeAfterEach:
				c.PushAfterEachNode(node)
			case types.ExampleComponentTypeJustBeforeEach:
				c.PushJustBeforeEachNode(node)
			}
		}
		return c
	}

	containers := func(containers ...*containernode.ContainerNode) []*containernode.ContainerNode {
		return containers
	}

	BeforeEach(func() {
		failer = Failer.New()
		codeLocation = codelocation.New(0)
		nodesThatRan = []string{}
	})

	Describe("marking examples focused and pending", func() {
		It("should satisfy various caes", func() {
			cases := []struct {
				ContainerFlags []internaltypes.FlagType
				SubjectFlag    internaltypes.FlagType
				Pending        bool
				Focused        bool
			}{
				{[]internaltypes.FlagType{}, noneFlag, false, false},
				{[]internaltypes.FlagType{}, focusedFlag, false, true},
				{[]internaltypes.FlagType{}, pendingFlag, true, false},
				{[]internaltypes.FlagType{noneFlag}, noneFlag, false, false},
				{[]internaltypes.FlagType{focusedFlag}, noneFlag, false, true},
				{[]internaltypes.FlagType{pendingFlag}, noneFlag, true, false},
				{[]internaltypes.FlagType{noneFlag}, focusedFlag, false, true},
				{[]internaltypes.FlagType{focusedFlag}, focusedFlag, false, true},
				{[]internaltypes.FlagType{pendingFlag}, focusedFlag, true, true},
				{[]internaltypes.FlagType{noneFlag}, pendingFlag, true, false},
				{[]internaltypes.FlagType{focusedFlag}, pendingFlag, true, true},
				{[]internaltypes.FlagType{pendingFlag}, pendingFlag, true, false},
				{[]internaltypes.FlagType{focusedFlag, noneFlag}, noneFlag, false, true},
				{[]internaltypes.FlagType{noneFlag, focusedFlag}, noneFlag, false, true},
				{[]internaltypes.FlagType{pendingFlag, noneFlag}, noneFlag, true, false},
				{[]internaltypes.FlagType{noneFlag, pendingFlag}, noneFlag, true, false},
				{[]internaltypes.FlagType{focusedFlag, pendingFlag}, noneFlag, true, true},
			}

			for i, c := range cases {
				subject := newIt("it node", c.SubjectFlag, false)
				containers := []*containernode.ContainerNode{}
				for _, flag := range c.ContainerFlags {
					containers = append(containers, newContainer("container", flag))
				}

				example := New(subject, containers)
				Ω(example.Pending()).Should(Equal(c.Pending), "Case %d: %#v", i, c)
				Ω(example.Focused()).Should(Equal(c.Focused), "Case %d: %#v", i, c)

				if c.Pending {
					Ω(example.Summary("").State).Should(Equal(types.ExampleStatePending))
				}
			}
		})
	})

	Describe("Skip", func() {
		It("should be skipped", func() {
			example := New(newIt("it node", noneFlag, false), containers(newContainer("container", noneFlag)))
			Ω(example.Skipped()).Should(BeFalse())
			example.Skip()
			Ω(example.Skipped()).Should(BeTrue())
			Ω(example.Summary("").State).Should(Equal(types.ExampleStateSkipped))
		})
	})

	Describe("IsMeasurement", func() {
		It("should be true if the subject is a measurement node", func() {
			example := New(newIt("it node", noneFlag, false), containers(newContainer("container", noneFlag)))
			Ω(example.IsMeasurement()).Should(BeFalse())
			Ω(example.Summary("").IsMeasurement).Should(BeFalse())
			Ω(example.Summary("").NumberOfSamples).Should(Equal(1))

			example = New(newMeasure("measure node", noneFlag, false, 10), containers(newContainer("container", noneFlag)))
			Ω(example.IsMeasurement()).Should(BeTrue())
			Ω(example.Summary("").IsMeasurement).Should(BeTrue())
			Ω(example.Summary("").NumberOfSamples).Should(Equal(10))
		})
	})

	Describe("Passed", func() {
		It("should pass when the subject passed", func() {
			example := New(newIt("it node", noneFlag, false), containers())
			example.Run()

			Ω(example.Passed()).Should(BeTrue())
			Ω(example.Failed()).Should(BeFalse())
			Ω(example.Summary("").State).Should(Equal(types.ExampleStatePassed))
			Ω(example.Summary("").Failure).Should(BeZero())
		})
	})

	Describe("Failed", func() {
		It("should be failed if the failure was panic", func() {
			example := New(newItWithBody("panicky it", func() {
				panic("bam")
			}), containers())
			example.Run()
			Ω(example.Passed()).Should(BeFalse())
			Ω(example.Failed()).Should(BeTrue())
			Ω(example.Summary("").State).Should(Equal(types.ExampleStatePanicked))
			Ω(example.Summary("").Failure.Message).Should(Equal("Test Panicked"))
			Ω(example.Summary("").Failure.ForwardedPanic).Should(Equal("bam"))
		})

		It("should be failed if the failure was a timeout", func() {
			example := New(newItWithBody("sleepy it", func(done Done) {}), containers())
			example.Run()
			Ω(example.Passed()).Should(BeFalse())
			Ω(example.Failed()).Should(BeTrue())
			Ω(example.Summary("").State).Should(Equal(types.ExampleStateTimedOut))
			Ω(example.Summary("").Failure.Message).Should(Equal("Timed out"))
		})

		It("should be failed if the failure was... a failure", func() {
			example := New(newItWithBody("failing it", func() {
				failer.Fail("bam", codeLocation)
			}), containers())
			example.Run()
			Ω(example.Passed()).Should(BeFalse())
			Ω(example.Failed()).Should(BeTrue())
			Ω(example.Summary("").State).Should(Equal(types.ExampleStateFailed))
			Ω(example.Summary("").Failure.Message).Should(Equal("bam"))
		})
	})

	Describe("Concatenated string", func() {
		It("should concatenate the texts of the containers and the subject", func() {
			example := New(
				newIt("it node", noneFlag, false),
				containers(
					newContainer("outer container", noneFlag),
					newContainer("inner container", noneFlag),
				),
			)

			Ω(example.ConcatenatedString()).Should(Equal("outer container inner container it node"))
		})
	})

	Describe("running it examples", func() {
		Context("with just an it", func() {
			Context("that succeeds", func() {
				It("should run the it and report on its success", func() {
					example := New(newIt("it node", noneFlag, false), containers())
					example.Run()
					Ω(example.Passed()).Should(BeTrue())
					Ω(example.Failed()).Should(BeFalse())
					Ω(nodesThatRan).Should(Equal([]string{"it node"}))
				})
			})

			Context("that fails", func() {
				It("should run the it and report on its success", func() {
					example := New(newIt("it node", noneFlag, true), containers())
					example.Run()
					Ω(example.Passed()).Should(BeFalse())
					Ω(example.Failed()).Should(BeTrue())
					Ω(example.Summary("").Failure.Message).Should(Equal("it node"))
					Ω(nodesThatRan).Should(Equal([]string{"it node"}))
				})
			})
		})

		Context("with a full set of setup nodes", func() {
			var failingNodes map[string]bool

			BeforeEach(func() {
				failingNodes = map[string]bool{}
			})

			JustBeforeEach(func() {
				example = New(
					newIt("it node", noneFlag, failingNodes["it node"]),
					containers(
						newContainer("outer container", noneFlag,
							newBef("outer bef A", failingNodes["outer bef A"]),
							newBef("outer bef B", failingNodes["outer bef B"]),
							newJusBef("outer jusbef A", failingNodes["outer jusbef A"]),
							newJusBef("outer jusbef B", failingNodes["outer jusbef B"]),
							newAft("outer aft A", failingNodes["outer aft A"]),
							newAft("outer aft B", failingNodes["outer aft B"]),
						),
						newContainer("inner container", noneFlag,
							newBef("inner bef A", failingNodes["inner bef A"]),
							newBef("inner bef B", failingNodes["inner bef B"]),
							newJusBef("inner jusbef A", failingNodes["inner jusbef A"]),
							newJusBef("inner jusbef B", failingNodes["inner jusbef B"]),
							newAft("inner aft A", failingNodes["inner aft A"]),
							newAft("inner aft B", failingNodes["inner aft B"]),
						),
					),
				)
				example.Run()
			})

			Context("that all pass", func() {
				It("should walk through the nodes in the correct order", func() {
					Ω(example.Passed()).Should(BeTrue())
					Ω(example.Failed()).Should(BeFalse())
					Ω(nodesThatRan).Should(Equal([]string{
						"outer bef A",
						"outer bef B",
						"inner bef A",
						"inner bef B",
						"outer jusbef A",
						"outer jusbef B",
						"inner jusbef A",
						"inner jusbef B",
						"it node",
						"inner aft A",
						"inner aft B",
						"outer aft A",
						"outer aft B",
					}))
				})
			})

			Context("when the subject fails", func() {
				BeforeEach(func() {
					failingNodes["it node"] = true
				})

				It("should run the afters", func() {
					Ω(example.Passed()).Should(BeFalse())
					Ω(example.Failed()).Should(BeTrue())
					Ω(nodesThatRan).Should(Equal([]string{
						"outer bef A",
						"outer bef B",
						"inner bef A",
						"inner bef B",
						"outer jusbef A",
						"outer jusbef B",
						"inner jusbef A",
						"inner jusbef B",
						"it node",
						"inner aft A",
						"inner aft B",
						"outer aft A",
						"outer aft B",
					}))
					Ω(example.Summary("").Failure.Message).Should(Equal("it node"))
				})
			})

			Context("when an inner before fails", func() {
				BeforeEach(func() {
					failingNodes["inner bef A"] = true
				})

				It("should not run any other befores, but it should run the subsequent afters", func() {
					Ω(example.Passed()).Should(BeFalse())
					Ω(example.Failed()).Should(BeTrue())
					Ω(nodesThatRan).Should(Equal([]string{
						"outer bef A",
						"outer bef B",
						"inner bef A",
						"inner aft A",
						"inner aft B",
						"outer aft A",
						"outer aft B",
					}))
					Ω(example.Summary("").Failure.Message).Should(Equal("inner bef A"))
				})
			})

			Context("when an outer before fails", func() {
				BeforeEach(func() {
					failingNodes["outer bef B"] = true
				})

				It("should not run any other befores, but it should run the subsequent afters", func() {
					Ω(example.Passed()).Should(BeFalse())
					Ω(example.Failed()).Should(BeTrue())
					Ω(nodesThatRan).Should(Equal([]string{
						"outer bef A",
						"outer bef B",
						"outer aft A",
						"outer aft B",
					}))
					Ω(example.Summary("").Failure.Message).Should(Equal("outer bef B"))
				})
			})

			Context("when an after fails", func() {
				BeforeEach(func() {
					failingNodes["inner aft B"] = true
				})

				It("should run all other afters, but mark the test as failed", func() {
					Ω(example.Passed()).Should(BeFalse())
					Ω(example.Failed()).Should(BeTrue())
					Ω(nodesThatRan).Should(Equal([]string{
						"outer bef A",
						"outer bef B",
						"inner bef A",
						"inner bef B",
						"outer jusbef A",
						"outer jusbef B",
						"inner jusbef A",
						"inner jusbef B",
						"it node",
						"inner aft A",
						"inner aft B",
						"outer aft A",
						"outer aft B",
					}))
					Ω(example.Summary("").Failure.Message).Should(Equal("inner aft B"))
				})
			})

			Context("when a just before each fails", func() {
				BeforeEach(func() {
					failingNodes["outer jusbef B"] = true
				})

				It("should run the afters, but not the subject", func() {
					Ω(example.Passed()).Should(BeFalse())
					Ω(example.Failed()).Should(BeTrue())
					Ω(nodesThatRan).Should(Equal([]string{
						"outer bef A",
						"outer bef B",
						"inner bef A",
						"inner bef B",
						"outer jusbef A",
						"outer jusbef B",
						"inner aft A",
						"inner aft B",
						"outer aft A",
						"outer aft B",
					}))
					Ω(example.Summary("").Failure.Message).Should(Equal("outer jusbef B"))
				})
			})

			Context("when an after fails after an earlier node has failed", func() {
				BeforeEach(func() {
					failingNodes["it node"] = true
					failingNodes["inner aft B"] = true
				})

				It("should record the earlier failure", func() {
					Ω(example.Passed()).Should(BeFalse())
					Ω(example.Failed()).Should(BeTrue())
					Ω(nodesThatRan).Should(Equal([]string{
						"outer bef A",
						"outer bef B",
						"inner bef A",
						"inner bef B",
						"outer jusbef A",
						"outer jusbef B",
						"inner jusbef A",
						"inner jusbef B",
						"it node",
						"inner aft A",
						"inner aft B",
						"outer aft A",
						"outer aft B",
					}))
					Ω(example.Summary("").Failure.Message).Should(Equal("it node"))
				})
			})
		})
	})

	Describe("running measurement examples", func() {
		Context("when the measurement succeeds", func() {
			It("should run N samples", func() {
				example = New(
					newMeasure("measure node", noneFlag, false, 3),
					containers(
						newContainer("container", noneFlag,
							newBef("bef A", false),
							newJusBef("jusbef A", false),
							newAft("aft A", false),
						),
					),
				)
				example.Run()

				Ω(example.Passed()).Should(BeTrue())
				Ω(example.Failed()).Should(BeFalse())
				Ω(nodesThatRan).Should(Equal([]string{
					"bef A",
					"jusbef A",
					"measure node",
					"aft A",
					"bef A",
					"jusbef A",
					"measure node",
					"aft A",
					"bef A",
					"jusbef A",
					"measure node",
					"aft A",
				}))
			})
		})

		Context("when the measurement fails", func() {
			It("should bail after the failure occurs", func() {
				example = New(
					newMeasure("measure node", noneFlag, true, 3),
					containers(
						newContainer("container", noneFlag,
							newBef("bef A", false),
							newJusBef("jusbef A", false),
							newAft("aft A", false),
						),
					),
				)
				example.Run()

				Ω(example.Passed()).Should(BeFalse())
				Ω(example.Failed()).Should(BeTrue())
				Ω(nodesThatRan).Should(Equal([]string{
					"bef A",
					"jusbef A",
					"measure node",
					"aft A",
				}))
			})
		})
	})

	Describe("Summary", func() {
		var (
			subjectCodeLocation        types.CodeLocation
			outerContainerCodeLocation types.CodeLocation
			innerContainerCodeLocation types.CodeLocation
			summary                    *types.ExampleSummary
		)

		BeforeEach(func() {
			subjectCodeLocation = codelocation.New(0)
			outerContainerCodeLocation = codelocation.New(0)
			innerContainerCodeLocation = codelocation.New(0)

			example = New(
				leafnodes.NewItNode("it node", func() {
					time.Sleep(10 * time.Millisecond)
				}, noneFlag, subjectCodeLocation, 0, failer, 0),
				containers(
					containernode.New("outer container", noneFlag, outerContainerCodeLocation),
					containernode.New("inner container", noneFlag, innerContainerCodeLocation),
				),
			)

			example.Run()
			Ω(example.Passed()).Should(BeTrue())
			summary = example.Summary("suite id")
		})

		It("should have the suite id", func() {
			Ω(summary.SuiteID).Should(Equal("suite id"))
		})

		It("should have the component texts and code locations", func() {
			Ω(summary.ComponentTexts).Should(Equal([]string{"outer container", "inner container", "it node"}))
			Ω(summary.ComponentCodeLocations).Should(Equal([]types.CodeLocation{outerContainerCodeLocation, innerContainerCodeLocation, subjectCodeLocation}))
		})

		It("should have a runtime", func() {
			Ω(summary.RunTime).Should(BeNumerically(">=", 10*time.Millisecond))
		})

		It("should not be a measurement, or have a measurement summary", func() {
			Ω(summary.IsMeasurement).Should(BeFalse())
			Ω(summary.Measurements).Should(BeEmpty())
		})
	})

	Describe("Summaries for measurements", func() {
		var summary *types.ExampleSummary

		BeforeEach(func() {
			example = New(leafnodes.NewMeasureNode("measure node", func(b Benchmarker) {
				b.RecordValue("a value", 7, "some info")
			}, noneFlag, codeLocation, 4, failer, 0), containers())
			example.Run()
			Ω(example.Passed()).Should(BeTrue())
			summary = example.Summary("suite id")
		})

		It("should include the number of samples", func() {
			Ω(summary.NumberOfSamples).Should(Equal(4))
		})

		It("should be a measurement", func() {
			Ω(summary.IsMeasurement).Should(BeTrue())
		})

		It("should have the measurements report", func() {
			Ω(summary.Measurements).Should(HaveKey("a value"))

			report := summary.Measurements["a value"]
			Ω(report.Name).Should(Equal("a value"))
			Ω(report.Info).Should(Equal("some info"))
			Ω(report.Results).Should(Equal([]float64{7, 7, 7, 7}))
		})
	})
})