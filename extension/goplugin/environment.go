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

package goplugin

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"plugin"
	"reflect"
	"sort"
	"strings"
	"time"

	gohan_db "github.com/cloudwan/gohan/db"
	"github.com/cloudwan/gohan/extension"
	"github.com/cloudwan/gohan/extension/goext"
	logger "github.com/cloudwan/gohan/log"
	"github.com/cloudwan/gohan/schema"
	"github.com/cloudwan/gohan/server/middleware"
	gohan_sync "github.com/cloudwan/gohan/sync"
	"github.com/twinj/uuid"
)

var log = logger.NewLogger()

// Handler is a generic handler
type Handler func(context goext.Context, environment goext.IEnvironment) error

// Handlers is a list of generic handlers
type Handlers []Handler

// PrioritizedHandlers is a prioritized list of generic handlers
type PrioritizedHandlers map[goext.Priority]Handlers

// EventPrioritizedHandlers is a per-event prioritized list of generic handlers
type EventPrioritizedHandlers map[string]PrioritizedHandlers

// SchemaHandler is a schema handler
type SchemaHandler func(context goext.Context, resource goext.Resource, environment goext.IEnvironment) error

// SchemaHandlers is a list of schema handlers
type SchemaHandlers []SchemaHandler

// PrioritizedSchemaHandlers is a prioritized list of schema handlers
type PrioritizedSchemaHandlers map[goext.Priority]SchemaHandlers

// SchemaPrioritizedSchemaHandlers is a per-schema prioritized list of schema handlers
type SchemaPrioritizedSchemaHandlers map[string]PrioritizedSchemaHandlers

// EventSchemaPrioritizedSchemaHandlers is a per-event per-schema prioritized list of schema handlers
type EventSchemaPrioritizedSchemaHandlers map[string]SchemaPrioritizedSchemaHandlers

// Environment golang based environment for gohan extensions
type Environment struct {
	initFns         map[string]func(goext.IEnvironment) error
	beforeStartHook func() error
	afterStopHook   func()

	extCore     goext.ICore
	extLogger   goext.ILogger
	extSchemas  goext.ISchemas
	extSync     goext.ISync
	extDatabase goext.IDatabase

	name  string
	db    gohan_db.DB
	ident middleware.IdentityService
	sync  gohan_sync.Sync

	traceID string

	handlers       EventPrioritizedHandlers
	schemaHandlers EventSchemaPrioritizedSchemaHandlers

	rawTypes map[string]reflect.Type
	types    map[string]reflect.Type
}

// NewEnvironment create new gohan extension rawEnvironment based on context
func NewEnvironment(name string, beforeStartHook func() error, afterStopHook func(), db gohan_db.DB, ident middleware.IdentityService, sync gohan_sync.Sync) *Environment {
	env := &Environment{
		initFns:         map[string]func(goext.IEnvironment) error{},
		beforeStartHook: beforeStartHook,
		afterStopHook:   afterStopHook,

		name:  name,
		db:    db,
		ident: ident,
		sync:  sync,

		rawTypes: make(map[string]reflect.Type),
		types:    make(map[string]reflect.Type),
	}
	env.bind()
	return env
}

// Core returns an implementation to Core interface
func (env *Environment) Core() goext.ICore {
	return env.extCore
}

// Logger returns an implementation to Logger interface
func (env *Environment) Logger() goext.ILogger {
	return env.extLogger
}

// Schemas returns an implementation to Schemas interface
func (env *Environment) Schemas() goext.ISchemas {
	return env.extSchemas
}

// Sync returns an implementation to Sync interface
func (env *Environment) Sync() goext.ISync {
	return env.extSync
}

// Database returns an implementation to IDatabase interface
func (env *Environment) Database() goext.IDatabase {
	return env.extDatabase
}

// Http returns an implementation to IHttp interface
func (thisEnvironment *Environment) Http() goext.IHttp {
	return &Http{}
}

//bind sets environment bindings
func (env *Environment) bind() {
	env.extCore = NewCore(env)
	env.extLogger = NewLogger(env)
	env.extSchemas = NewSchemas(env)
	env.extSync = NewSync(env)
	env.extDatabase = NewDatabase(env)
}

// Start starts already loaded environment
func (env *Environment) Start() error {
	var err error

	if len(env.initFns) == 0 {
		log.Debug("Go environment is empty: %s", env.name)
		return nil
	}

	log.Debug("Starting go environment: %s", env.name)

	// Before start init
	if env.beforeStartHook != nil {
		log.Debug("Calling before start for go environment: %s", env.name)

		if err = env.beforeStartHook(); err != nil {
			log.Error("Failed to call before start for go extension: %s; error: %s", env.name, err)
			return err
		}
	} else {
		log.Debug("Before start init is not set for go environment: %s", env.name)
	}

	// Manager
	//env.manager = schema.GetManager()

	// bind
	env.bind()

	// generate TraceID
	env.traceID = uuid.NewV4().String()

	// get init
	log.Debug("Start go extension: %s", env.name)

	for _, initFn := range env.initFns {
		err = initFn(env)

		if err != nil {
			log.Error("Failed to start go extension: %s; error: %s", env.name, err)
			return err
		}
	}

	log.Debug("Go extension started: %s", env.name)

	return nil
}

// Load loads script into the environment
func (env *Environment) Load(binary string) (bool, error) {
	log.Debug("Loading go extension: %s", binary)

	if _, ok := env.initFns[binary]; ok {
		log.Warning("Go extension %s already loaded in %s", binary, env.name)
		return true, nil
	}

	var err error
	var ok bool

	if filepath.Ext(binary) != ".so" {
		return false, fmt.Errorf("go extension must be a *.so file, file: %s", binary)
	}

	pl, err := plugin.Open(binary)

	if err != nil {
		return false, fmt.Errorf("failed to load go extension: %s", err)
	}

	// Init
	initFnRaw, err := pl.Lookup("Init")

	if err != nil {
		return false, fmt.Errorf("go extension does not export Init: %s", err)
	}

	initFn, ok := initFnRaw.(func(goext.IEnvironment) error)

	if !ok {
		return false, fmt.Errorf("invalid signature of Init function in go extension: %s", binary)
	}

	env.initFns[binary] = initFn

	return true, nil
}

//LoadExtensionsForPath for returns extensions for specific path
func (env *Environment) LoadExtensionsForPath(extensions []*schema.Extension, timeLimit time.Duration, timeLimits []*schema.PathEventTimeLimit, path string) error {
	for _, extension := range extensions {
		if extension.Match(path) {
			if extension.CodeType != "goext" {
				continue
			}
			url := strings.TrimPrefix(extension.URL, "file://")
			if url == "" {
				log.Warning("ignore go extension '%s' without binary", extension.ID)
				continue
			}
			loaded, err := env.Load(url)
			if err != nil {
				return err
			}
			if loaded {
				if err = env.Start(); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (env *Environment) dispatchSchemaEvent(prioritizedSchemaHandlers PrioritizedSchemaHandlers, sch Schema, event string, context map[string]interface{}) error {
	env.Logger().Debugf("Starting event: %s, schema: %s", event, sch.raw.ID)
	defer env.Logger().Debugf("Finished event: %s, schema: %s", event, sch.raw.ID)
	if resource, err := env.resourceFromContext(sch, context); err == nil {
		for _, priority := range sortSchemaHandlers(prioritizedSchemaHandlers) {
			for _, schemaEventHandler := range prioritizedSchemaHandlers[priority] {
				context["go_validation"] = true
				if err := schemaEventHandler(context, resource, env); err != nil {
					return err
				}
				context["resource"] = sch.StructToMap(resource)
			}
		}
	} else {
		return goext.Error{goext.ErrorBadRequest, fmt.Errorf("failed to parse resource from context with schema '%s' for event '%s': %s", sch.ID(), event, err)}
	}

	return nil
}

func sortSchemaHandlers(schemaHandlers PrioritizedSchemaHandlers) []goext.Priority {
	priorities := []goext.Priority{}
	for priority := range schemaHandlers {
		priorities = append(priorities, priority)
	}
	sort.Ints(priorities)
	return priorities
}

func sortHandlers(handlers PrioritizedHandlers) []goext.Priority {
	priorities := []goext.Priority{}
	for priority := range handlers {
		priorities = append(priorities, priority)
	}
	sort.Ints(priorities)
	return priorities
}

// HandleEvent handles an event
func (env *Environment) HandleEvent(event string, context map[string]interface{}) error {
	context["event_type"] = event
	// dispatch to schema handlers
	if schemaPrioritizedSchemaHandlers, ok := env.schemaHandlers[event]; ok {
		if iSchemaID, ok := context["schema_id"]; ok {
			schemaID := iSchemaID.(string)
			if prioritizedSchemaHandlers, ok := schemaPrioritizedSchemaHandlers[schemaID]; ok {
				if iSchema := env.Schemas().Find(schemaID); iSchema != nil {
					sch := iSchema.(*Schema)
					if err := env.dispatchSchemaEvent(prioritizedSchemaHandlers, *sch, event, context); err != nil {
						return err
					}
				}
			}
		} else {
			// all
			for schemaID, prioritizedSchemaHandlers := range schemaPrioritizedSchemaHandlers {
				if iSchema := env.Schemas().Find(schemaID); iSchema != nil {
					sch := iSchema.(*Schema)
					if err := env.dispatchSchemaEvent(prioritizedSchemaHandlers, *sch, event, context); err != nil {
						return err
					}
				} else {
					return fmt.Errorf("could not find schema: %s", schemaID)
				}
			}
		}
	}

	// dispatch to generic handlers
	if prioritizedEventHandlers, ok := env.handlers[event]; ok {
		for _, priority := range sortHandlers(prioritizedEventHandlers) {
			for index, eventHandler := range prioritizedEventHandlers[priority] {
				if err := eventHandler(context, env); err != nil {
					return fmt.Errorf("failed to dispatch event '%s' at priority '%d' with index '%d': %s",
						event, priority, index, err)
				}
			}
		}
	}

	return nil
}

func (env *Environment) updateContextFromResource(context goext.Context, resource interface{}) error {
	if resource == nil {
		context["resource"] = nil
		return nil
	}

	if _, ok := context["resource"]; !ok {
		return nil
	}

	if _, ok := context["resource"].(map[string]interface{}); !ok {
		return fmt.Errorf("failed to convert context resource to map during update context from resource")
	}

	if resourceMap, ok := env.resourceToMap(resource).(map[string]interface{}); ok {
		for key, value := range resourceMap {
			if _, ok := context["resource"].(map[string]interface{})[key]; ok {
				context["resource"].(map[string]interface{})[key] = value
			}
		}
	} else {
		return fmt.Errorf("failed to convert resource to map during update context from resource")
	}

	return nil
}

func (env *Environment) updateResourceFromContextR(resource interface{}, resourceData map[string]interface{}) error {
	resourceValue := reflect.ValueOf(resource)
	resourceElem := resourceValue.Elem()
	resourceElemType := resourceElem.Type()

	if resourceElemType.Kind() != reflect.Struct {
		panic("resource must be a struct")
	}

	for i := 0; i < resourceElemType.NumField(); i++ {
		resourceFieldType := resourceElemType.Field(i)
		resourceFieldTagDB := resourceFieldType.Tag.Get("db")
		resourceField := resourceElem.Field(i)
		val := reflect.ValueOf(resourceData[resourceFieldTagDB])

		if resourceFieldType.Type.Kind() == reflect.Struct {
			if _, ok := resourceData[resourceFieldTagDB].(map[string]interface{}); ok {
				env.updateResourceFromContextR(resourceField.Interface(), resourceData[resourceFieldTagDB].(map[string]interface{}))
			} else if strings.Contains(resourceFieldType.Type.String(), "goext.Null") {
				if resourceData[resourceFieldTagDB] != nil {
					if val.Type() == resourceFieldType.Type {
						resourceField.Set(val)
					} else {
						resourceField.Field(0).Set(val)
						resourceField.Field(1).Set(reflect.ValueOf(true))
					}
				} else {
					resourceField.Field(1).Set(reflect.ValueOf(false))
				}
			} else {
				resourceField.Set(val)
			}
		} else {
			if val.IsValid() {
				resourceField.Set(val)
			}
		}
	}

	return nil
}

func (env *Environment) updateResourceFromContext(resource interface{}, context goext.Context) error {
	if resource == nil {
		return nil
	}

	if _, ok := context["resource"]; !ok {
		return nil
	}

	if resourceData, ok := context["resource"].(map[string]interface{}); ok {
		return env.updateResourceFromContextR(resource, resourceData)
	}

	return fmt.Errorf("failed to convert context resource to map during update resource from context")
}

func (env *Environment) resourceToMap(resource interface{}) interface{} {
	resourceValue := reflect.ValueOf(resource)
	resourceElem := resourceValue.Elem()
	resourceElemType := resourceElem.Type()

	if resourceElemType.Kind() == reflect.Struct {
		switch res := resource.(type) {
		case *goext.NullString:
			if res.Valid {
				return res.Value
			}
			return nil
		case *goext.NullInt:
			if res.Valid {
				return res.Value
			}
			return nil
		case *goext.NullFloat:
			if res.Valid {
				return res.Value
			}
			return nil
		case *goext.NullBool:
			if res.Valid {
				return res.Value
			}
			return nil
		case goext.NullString:
			if res.Valid {
				return res.Value
			}
			return nil
		case goext.NullInt:
			if res.Valid {
				return res.Value
			}
			return nil
		case goext.NullFloat:
			if res.Valid {
				return res.Value
			}
			return nil
		case goext.NullBool:
			if res.Valid {
				return res.Value
			}
			return nil
		}
		data := make(map[string]interface{})

		for i := 0; i < resourceElemType.NumField(); i++ {
			resourceFieldType := resourceElemType.Field(i)
			resourceFieldTagDB := resourceFieldType.Tag.Get("db")
			resourceFieldInterface := resourceElem.Field(i).Interface()

			data[resourceFieldTagDB] = env.resourceToMap(&resourceFieldInterface)
		}

		return data
	}

	return resourceElem.Interface()
}

func (env *Environment) resourceFromContext(sch Schema, context map[string]interface{}) (res goext.Resource, err error) {
	rawSchema := sch.raw

	resourceType, ok := env.rawTypes[rawSchema.ID]
	if !ok {
		return nil, fmt.Errorf("No type registered for title: %s", rawSchema.ID)
	}

	resource := reflect.New(resourceType)

	resourceData, ok := context["resource"]

	if ok {
		data := resourceData.(map[string]interface{})
		for i := 0; i < resourceType.NumField(); i++ {
			field := resource.Elem().Field(i)
			fieldType := resourceType.Field(i)
			propertyName := fieldType.Tag.Get("db")
			if propertyName == "" {
				return nil, fmt.Errorf("Missing tag 'db' for resource %s field %s", resourceType.Name(), fieldType.Name)
			}
			property, err := rawSchema.GetPropertyByID(propertyName)
			if err != nil {
				return nil, err
			}
			kind := fieldType.Type.Kind()
			if kind == reflect.Struct || kind == reflect.Ptr || kind == reflect.Slice {
				mapJSON, err := json.Marshal(data[property.ID])
				if err != nil {
					return nil, err
				}
				newField := reflect.New(field.Type())
				fieldJSON := string(mapJSON)
				fieldInterface := newField.Interface()
				err = json.Unmarshal([]byte(fieldJSON), &fieldInterface)
				if err != nil {
					return nil, err
				}
				field.Set(newField.Elem())
			} else {
				value := reflect.ValueOf(data[property.ID])
				if value.IsValid() {
					if value.Type() != field.Type() && field.Kind() == reflect.Int && value.Kind() == reflect.Float64 { // reflect treats number(N, 0) as float
						field.SetInt(int64(value.Float()))
					} else {
						field.Set(value)
					}
				}
			}

		}
	}

	return resource.Interface(), nil
}

// RegisterEventHandler registers an event handler
func (env *Environment) RegisterEventHandler(event string, handler func(context goext.Context, environment goext.IEnvironment) error, priority goext.Priority) {
	if env.handlers == nil {
		env.handlers = EventPrioritizedHandlers{}
	}

	if env.handlers[event] == nil {
		env.handlers[event] = PrioritizedHandlers{}
	}

	if env.handlers[event][priority] == nil {
		env.handlers[event][priority] = Handlers{}
	}

	env.handlers[event][priority] = append(env.handlers[event][priority], handler)
}

// RegisterSchemaEventHandler register an event handler for a schema
func (env *Environment) RegisterSchemaEventHandler(schemaID string, event string, handler func(context goext.Context, resource goext.Resource, environment goext.IEnvironment) error, priority goext.Priority) {
	if env.schemaHandlers == nil {
		env.schemaHandlers = EventSchemaPrioritizedSchemaHandlers{}
	}

	if env.schemaHandlers[event] == nil {
		env.schemaHandlers[event] = SchemaPrioritizedSchemaHandlers{}
	}

	if env.schemaHandlers[event][schemaID] == nil {
		env.schemaHandlers[event][schemaID] = PrioritizedSchemaHandlers{}
	}

	if env.schemaHandlers[event][schemaID][priority] == nil {
		env.schemaHandlers[event][schemaID][priority] = SchemaHandlers{}
	}

	env.schemaHandlers[event][schemaID][priority] = append(env.schemaHandlers[event][schemaID][priority], handler)
}

// RegisterRawType registers a runtime type of raw resource for a given name
func (env *Environment) RegisterRawType(name string, typeValue interface{}) {
	targetType := reflect.TypeOf(typeValue)
	env.rawTypes[name] = targetType
}

// RawType returns a runtime type for a given named raw resource
func (env *Environment) RawType(name string) reflect.Type {
	return env.rawTypes[name]
}

// RegisterType registers a runtime type of resource for a given name
func (env *Environment) RegisterType(name string, typeValue interface{}) {
	targetType := reflect.TypeOf(typeValue)
	env.types[name] = targetType
}

// ResourceType returns a runtime type for a given named resource
func (env *Environment) ResourceType(name string) reflect.Type {
	return env.types[name]
}

// Stop stops the environment to its initial state
func (env *Environment) Stop() {
	log.Info("Stop environment")

	// reset locals
	env.extCore = nil
	env.extLogger = nil
	env.extSchemas = nil
	env.extSync = nil
	env.extDatabase = nil

	// after stop
	if env.afterStopHook != nil {
		log.Debug("Calling after stop hook for go environment: %s", env.name)
		env.afterStopHook()
	} else {
		log.Debug("After stop hook is not set for go environment: %s", env.name)
	}
}

// Reset clear the environment to its initial state
func (env *Environment) Reset() {
	log.Info("Reset environment")

	env.Stop()
	env.Start()
}

// Clone makes a clone of the rawEnvironment
func (env *Environment) Clone() extension.Environment {
	clone := &Environment{
		initFns:         env.initFns,
		beforeStartHook: env.beforeStartHook,
		afterStopHook:   env.afterStopHook,

		name:  env.name,
		db:    env.db,
		ident: env.ident,
		sync:  env.sync,

		traceID: uuid.NewV4().String(),

		rawTypes: env.rawTypes,
		types:    env.types,
	}
	clone.bind()
	return clone
}

// IsEventHandled returns whether a given event is handled by this environment
func (env *Environment) IsEventHandled(event string, context map[string]interface{}) bool {
	return true
}