package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/vmware-tanzu/octant/pkg/navigation"
	"github.com/vmware-tanzu/octant/pkg/plugin"
	"github.com/vmware-tanzu/octant/pkg/plugin/service"
	"github.com/vmware-tanzu/octant/pkg/store"
	"github.com/vmware-tanzu/octant/pkg/view/component"
	"github.com/vmware-tanzu/octant/pkg/view/flexlayout"
)

var pluginName = "octant-riff"

// This is a sample plugin showing the features of Octant's plugin API.
func main() {
	// Remove the prefix from the go logger since Octant will print logs with timestamps.
	log.SetPrefix("")

	// This plugin is interested in Pods
	functionGVK := schema.GroupVersionKind{Group: "build.projectriff.io", Version: "v1alpha1", Kind: "Function"}

	// Tell Octant to call this plugin when printing configuration or tabs for Pods
	capabilities := &plugin.Capabilities{
		SupportsPrinterConfig: []schema.GroupVersionKind{functionGVK},
		SupportsTab:           []schema.GroupVersionKind{functionGVK},
		IsModule:              true,
	}

	// Set up what should happen when Octant calls this plugin.
	options := []service.PluginOption{
		service.WithPrinter(handlePrint),
		service.WithTabPrinter(handleTab),
		service.WithNavigation(handleNavigation, initRoutes),
	}

	// Use the plugin service helper to register this plugin.
	p, err := service.Register(pluginName, "visualize riff", capabilities, options...)
	if err != nil {
		log.Fatal(err)
	}

	// The plugin can log and the log messages will show up in Octant.
	log.Printf("octant-riff plugin is starting")
	//fmt.Println("octant-riff plugin is starting")
	p.Serve()
}

// handleTab is called when Octant wants to print a tab for an object.
func handleTab(request *service.PrintRequest) (plugin.TabResponse, error) {
	log.Println("SWAP:handleTab")
	if request.Object == nil {
		return plugin.TabResponse{}, errors.New("object is nil")
	}

	// Octant uses flex layouts to display information. It's a flexible
	// grid. A flex layout is composed of multiple section. Each section
	// can contain multiple components. Components are displayed given
	// a width. In the case below, the width is half of the visible space.
	// Create sections to separate your components as each section will
	// start a new row.
	layout := flexlayout.New()
	section := layout.AddSection()

	// Octant contain's a library of components that can be used to display content.
	// This example uses markdown text.
	contents := component.NewMarkdownText("content from a *plugin*")

	err := section.Add(contents, component.WidthHalf)
	if err != nil {
		return plugin.TabResponse{}, err
	}

	// In this example, this plugin will tell Octant to create a new
	// tab when showing pods. This tab's name will be "Extra Pod Details".
	tab := component.NewTabWithContents(*layout.ToComponent("Function Details"))

	return plugin.TabResponse{Tab: tab}, nil
}

// handlePrint is called when Octant wants to print an object.
func handlePrint(request *service.PrintRequest) (plugin.PrintResponse, error) {
	log.Println("SWAP:handlePrint")
	if request.Object == nil {
		return plugin.PrintResponse{}, errors.Errorf("object is nil")
	}

	// load an object from the cluster and use that object to create a response.

	// Octant has a helper function to generate a key from an object. The key
	// is used to find the object in the cluster.
	key, err := store.KeyFromObject(request.Object)
	if err != nil {
		return plugin.PrintResponse{}, err
	}
	u, err := request.DashboardClient.Get(request.Context(), key)
	if err != nil {
		return plugin.PrintResponse{}, err
	}

	// The plugin can check if the object it requested exists.
	if u == nil {
		return plugin.PrintResponse{}, errors.New("object doesn't exist")
	}

	// Octant has a component library that can be used to build content for a plugin.
	// In this case, the plugin is creating a card.
	podCard := component.NewCard(component.TitleFromString(fmt.Sprintf("List of processors for %s", u.GetName())))
	processorList, err := getProcessorsForFunction(request.DashboardClient, u)
	if err != nil {
		return plugin.PrintResponse{}, nil
	}
	podCard.SetBody(component.NewMarkdownText(processorList))

	msg := fmt.Sprintf("update from plugin at %s", time.Now().Format(time.RFC3339))

	// When printing an object, you can create multiple types of content. In this
	// example, the plugin is:
	//
	// * adding a field to the configuration section for this object.
	// * adding a field to the status section for this object.
	// * create a new piece of content that will be embedded in the
	//   summary section for the component.
	return plugin.PrintResponse{
		Config: []component.SummarySection{
			{Header: "from-plugin", Content: component.NewText(msg)},
		},
		Status: []component.SummarySection{
			{Header: "from-plugin", Content: component.NewText(msg)},
		},
		Items: []component.FlexLayoutItem{
			{
				Width: component.WidthHalf,
				View:  podCard,
			},
		},
	}, nil
}

func getFunctions(client service.Dashboard) []string {
	functionGVK := schema.GroupVersionKind{Group: "build.projectriff.io", Version: "v1alpha1", Kind: "Function"}
	functionKey := store.KeyFromGroupVersionKind(functionGVK)
	result := []string{}

	l, err := client.List(context.Background(), functionKey)
	if err != nil {
		panic(err)
		//return append(result, component.NewText(fmt.Sprintf("%s", err)))
	}

	for _, i := range l.Items {
		result = append(result, i.GetName())
	}
	//for _, i := range l.Items {
	//	result = append(result, component.NewLink(i.GetName(), "",
	//		fmt.Sprintf("/custom-resources/functions.build.projectriff.io/v1alpha1/%s", i.GetName())))
	//}
	return result
}

func getProcessorsForFunction(client service.Dashboard, function *unstructured.Unstructured) (string, error) {
	//processorGVK := schema.GroupVersionKind{Group: "streaming.projectriff.io", Version: "v1alpha1", Kind: "Processor"}
	//processorKey := store.KeyFromGroupVersionKind(processorGVK)
	//l, err := client.List(context.Background(), processorKey)
	//if err != nil {
	//	return "", err
	//}
	return "WORK IN PROGRESS", nil
}

// handlePrint creates a navigation tree for this plugin. Navigation is dynamic and will
// be called frequently from Octant. Navigation is a tree of `Navigation` structs.
// The plugin can use whatever paths it likes since these paths can be namespaced to the
// the plugin.
func handleNavigation(request *service.NavigationRequest) (navigation.Navigation, error) {
	return navigation.Navigation{
		Title: "riff Plugin",
		Path:  request.GeneratePath(),
		IconName: "cloud",
	}, nil
}

// initRoutes routes for this plugin. In this example, there is a global catch all route
// that will return the content for every single path.
func initRoutes(router *service.Router) {

	router.HandleFunc("*", func(request service.Request) (component.ContentResponse, error) {

		card := component.NewCard([]component.TitleComponent{component.NewText("functions")})

		processorGVK := schema.GroupVersionKind{Group: "streaming.projectriff.io", Version: "v1alpha1", Kind: "Processor"}
		processorKey := store.KeyFromGroupVersionKind(processorGVK)
		l, err := request.DashboardClient().List(context.Background(), processorKey)
		if err != nil {
			return component.ContentResponse{}, err
		}
		functions := map[string]bool{}
		table := component.NewTable("Functions", "placeholder",
			[]component.TableCol{
				component.TableCol{Name: "function"},
				component.TableCol{Name: "processor"},
				component.TableCol{Name: "input streams"},
				component.TableCol{Name: "output streams"},
			})

		component.new
		sortedList := sortUnstructuredList(l.Items)
		for _, i := range sortedList {
			funcName, found, err := unstructured.NestedString(i.UnstructuredContent(), "spec", "build", "functionRef")
			if err != nil || !found {
				funcName = "NOT FOUND"
			}
			inStreamStr := getStreamNames(i, "spec", "inputs")
			outStreamStr := getStreamNames(i, "spec", "outputs")
			table.Add(component.TableRow{
				"function":      component.NewText(funcName),
				"processor":     component.NewText(i.GetName()),
				"input streams": component.NewText(fmt.Sprintf("%s", inStreamStr)),
				"output streams": component.NewText(fmt.Sprintf("%s", outStreamStr)),
			})
			functions[funcName] = true
		}

		allFunctions := getFunctions(request.DashboardClient())
		for fi := range allFunctions {
			if _, ok := functions[allFunctions[fi]]; !ok {
				table.Add(component.TableRow{
					"function": component.NewText(allFunctions[fi]),
				})
			}
		}

		card.SetBody(table)
		contentResponse := component.NewContentResponse(component.TitleFromString("riff Components"))
		contentResponse.Add(card)

		return *contentResponse, nil
	})
}

func sortUnstructuredList(items []unstructured.Unstructured) []unstructured.Unstructured {
	sort.Slice(items, func(i, j int) bool {
		return items[i].GetName() < items[j].GetName()
	})
	return items
}

func getStreamNames(i unstructured.Unstructured, fields ...string) []string {
	inStreamStr := []string{}
	inStreams, found, err := unstructured.NestedSlice(i.UnstructuredContent(), fields...)
	if err != nil {
		return append(inStreamStr, fmt.Sprintf("%s", err))
	}
	if !found {
		return append(inStreamStr, "")
	}
	for i := range inStreams {
		stream := inStreams[i].(map[string]interface{})
		inStreamStr = append(inStreamStr, fmt.Sprintf("%s", stream["stream"]))
	}
	return inStreamStr
}
