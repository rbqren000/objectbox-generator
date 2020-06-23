<img width="466" src="https://raw.githubusercontent.com/objectbox/objectbox-java/master/logo.png">
<br/>

[![Follow ObjectBox on Twitter](https://img.shields.io/twitter/follow/ObjectBox_io.svg?style=flat-square&logo=twitter&color=fff)](https://twitter.com/ObjectBox_io)

# ObjectBox Generator

Current version: 0.9.1

ObjectBox is a superfast cross-platform object-oriented database.
ObjectBox Generator takes over the burden of writing the C/C++ binding code for ObjectBox (more languages to be supported in the future).
This greatly simplifies the model declaration and FlatBuffers serialization, allowing you to concentrate on the actual application logic.

All the generated code is header-only and compatible with the existing ObjectBox C-API so you can take advantage of the new features while incrementally porting your existing code.

## Prerequisites/Installation

1. You need to have [ObjectBox-C](https://github.com/objectbox/objectbox-c) library installed to use code generated by ObjectBox Generator in your project.
   Please follow the [installation instructions](https://github.com/objectbox/objectbox-c#usage-and-installation).

2. Install the objectbox-generator by downloading the latest binary for your OS from [releases](https://github.com/objectbox/objectbox-generator/releases/latest).
   If you want, add it to $PATH for convenience.
   Alternatively, instead of downloading, you can build the generator yourself by cloning this repo and running `make`.
   To build yourself, you need a recent Go version, CMake and a C++11 tool chain.

3. Get a FlatBuffers library:
    * For C: get [flatc library and headers](https://github.com/dvidelabs/flatcc)
    * For C++: get [flatbuffers headers](https://github.com/google/flatbuffers/tree/v1.12.0/include/flatbuffers).
      Note: objectbox-c library (.so/.dll) already includes this library.

## Getting started

ObjectBox Generator uses FlatBuffer schema file (.fbs) as its input.
It stores the model information within a model JSON file known from other ObjectBox language bindings (objectbox-model.json) and generates code based on the selected language - C or C++.

Let’s have a look at a sample schema (`tasklist.fbs`) and how Generator helps us.

```text
table Task {
    id: ulong;
    text: string;
    date_created: ulong;
    date_finished: ulong;
}
```

### For C++ projects

Running `objectbox-generator -cpp tasklist.fbs` will generate C++ binding code 
for `tasklist.fbs` - we get the following files:

* objectbox-model.h
* objectbox-model.json
* tasklist-cpp.obx.h

> Note: you should add all these files to your source control (e.g. git), 
> most importantly the objectbox-model.json which ensures compatibility 
> with previous versions of your database after you make changes to the schema.

Now in your application, you can include the headers and start to work with your database. 
Consider the following `main.cpp`:

```cpp
#include "objectbox-cpp.h"
#include "objectbox-model.h"
#include "tasklist-cpp.obx.h"

int main(int argc, char* args[]) {
    // create_obx_model() provided by objectbox-model.h
    // obx interface contents provided by objectbox-ext.h
    obx::Store store(create_obx_model());
    obx::Box<Task> box(store);

    obx_id id = box.put({.text = "Buy milk"});  // Create
    std::unique_ptr<Task> task = box.get(id);   // Read
    if (task) {
        task->text += " & some bread";
        box.put(*task);                         // Update
        box.remove(id);                         // Delete
    }
    return 0;
}
```

To compile, just link to the objectbox-c library, e.g. something like this should 
work: `g++ main.cpp -I. -std=c++11 -lobjectbox`. Note: the command snippet assumes 
you have objectbox-c library installed in a path recognized by your OS (e.g. /usr/local/lib/)
and all the referenced headers are in the same folder as `main.cpp`.

### For C projects

Running `objectbox-generator -c tasklist.fbs` will generate C binding code for 
`tasklist.fbs` - we get the following files:

* objectbox-model.h
* objectbox-model.json
* tasklist.obx.h

> Note: you should add all these files to your source control (e.g. git), 
> most importantly the objectbox-model.json which ensures compatibility 
> with previous versions of your database after you make changes to the schema.

Now in your application, you can include the headers and start to work with your database. 
Have a look at the following `main.c` showing one of the many ways you can work with 
objectbox-c and the generated code:

```c
#include "objectbox-model.h"
#include "objectbox.h"
#include "tasklist.obx.h"

obx_err print_last_error() {
    printf("Unexpected error: %d %s\n", obx_last_error_code(), obx_last_error_message());
    return obx_last_error_code();
}

obx_id task_put(OBX_box* box, Task* task) {
    flatcc_builder_t builder;
    flatcc_builder_init(&builder);

    size_t size = 0;
    void* buffer = NULL;

    // Note: Task_to_flatbuffer() is provided by the generated code
    obx_id id = 0;
    if (Task_to_flatbuffer(&builder, task, &buffer, &size)) {
        id = obx_box_put_object(box, buffer, size, OBXPutMode_PUT);  // returns 0 on error
    }

    flatcc_builder_clear(&builder);
    if (buffer) flatcc_builder_aligned_free(buffer);

    if (id == 0) {
        // TODO: won't be able to print the right error if it occurred in Task_to_flatbuffer(), 
        //  i.e. outside objectbox
        print_last_error();
    } else {
        task->id = id;  // Note: we're updating the ID on new objects for convenience
    }

    return id;
}

Task* task_read(OBX_store* store, OBX_box* box, obx_id id) {
    OBX_txn* txn = NULL;

    // We need an explicit TX - read flatbuffers lifecycle is bound to the open transaction.
    // The transaction can be closed safely after reading the object properties from flatbuffers.
    txn = obx_txn_read(store);
    if (!txn) {
        print_last_error();
        return NULL;
    }

    void* data;
    size_t size;
    int rc = obx_box_get(box, id, &data, &size);
    if (rc != OBX_SUCCESS) {
        // if (rc == OBX_NOT_FOUND); // No special treatment at the moment if not found
        obx_txn_close(txn);
        return NULL;
    }

    Task* result = Task_new_from_flatbuffer(data, size);
    obx_txn_close(txn);
    return result;
}

int main(int argc, char* args[]) {
    int rc = 0;
    OBX_store* store = NULL;
    OBX_box* box = NULL;
    Task* task = NULL;

    // Firstly, we need to create a model for our data and the store
    {
        OBX_model* model = create_obx_model();  // create_obx_model() provided by objectbox-model.h
        if (!model) goto handle_error;
        if (obx_model_error_code(model)) {
            printf("Model definition error: %d %s\n", 
                obx_model_error_code(model), obx_model_error_message(model));
            obx_model_free(model);
            goto handle_error;
        }

        OBX_store_options* opt = obx_opt();
        obx_opt_model(opt, model);
        store = obx_store_open(opt);
        if (!store) goto handle_error;

        // obx_store_open() takes ownership of model and opt and frees them.
    }

    box = obx_box(store, Task_ENTITY_ID);  // Note the generated "Task_ENTITY_ID"

    obx_id id = 0;

    {  // Create
        Task task = {.text = "Buy milk"};
        id = task_put(box, &task);
        if (!id) goto handle_error;
        printf("New task inserted with ID %d\n", id);
    }

    {  // Read
        task = task_read(store, box, id);
        if (!task) goto handle_error;
        printf("Task %d read with text: %s\n", id, task->text);
    }

    {  // Update
        const char* appendix = " & some bread";

        // updating a string property is a little more involved but nothing too hard
        size_t old_text_len = task->text ? strlen(task->text) : 0;
        char* new_text = (char*) malloc((old_text_len + strlen(appendix) + 1) * sizeof(char));

        if (task->text) {
            memcpy(new_text, task->text, old_text_len);

            // free the previously allocated memory or it would be lost when overwritten below
            free(task->text);
        }
        memcpy(new_text + old_text_len, appendix, strlen(appendix) + 1);
        task->text = new_text;
        printf("Updated task %d with a new text: %s\n", id, task->text);
    }

    // Delete
    if (obx_box_remove(box, id) != OBX_SUCCESS) goto handle_error;

free_resources:  // free any remaining allocated resources
    if (task) Task_free(&task); // We must free the object allocated by Task_new_from_flatbuffer() 
    if (store) obx_store_close(store); // And close the store. Boxes are closed automatically.
    return rc;

handle_error:  // print error and clean up
    rc = print_last_error();
    if (rc <= 0) rc = 1;
    goto free_resources;
}
```

To compile, link to the objectbox-c library and flatcc-runtime library, 
e.g. something like this should work: `gcc main.c -I. -lobjectbox -lflatccrt`. 
Note: the command snippet assumes you have objectbox-c and flatccrt libraries installed in a path 
recognized by your OS (e.g. /usr/local/lib/) and all the referenced headers are in the same folder as `main.c`.

## Annotations

The source FlatBuffer schema can contain some ObjectBox-specific annotations, declared as specially 
formatted comments to `table` and `field` FlatBuffer schema elements. Have a look at the following 
schema example showing of a few of the annotations and the various formats you can use.

```text
/// This entity is not annotated and only serves as a relation target in this example
table Simple {
    id:ulong;
}

/// objectbox: name=AnnotatedEntity
table Annotated {
    /// Objectbox requires an ID property.
    /// Recognized automatically if it has a right name ("id"), otherwise it must be annotated.
    /// objectbox:id
    identifier:ulong;

    /// objectbox:name="name",index=hash64
    fullName:string;

    /// objectbox:id-companion, date
    time:int64;

    /// objectbox:transient
    skippedField:[uint64];

    /// objectbox:link=Simple
    relId:ulong;
}
```

### Annotation format

The simplified rules how annotation-specific comments are recognized:

* Must be a comment immediately preceding an Entity or a Property (no empty lines between them).
* The comment must start with three slashes so it's be picked up by FlatBuffer schema parser as  "documentation".
* Spaces between words inside the comment are skipped so you can use them for better readability, if you like. See e.g. `Annotated`, `time`.
* The comment must start with the text `objectbox:` and is followed by one or more annotations, separatedy by commas.
* Each annotation has a name and some annotations also support specifying a value (some even require a value, e.g. `name`). See e.g. `Annotated`, `relId`.
* Value, if present, is added to the annotation by adding an equal sign and the actual value.
* A value may additionally be surrounded by double quotes but it's not necessary. See e.g. `fullName` showing both variants.

### Supported annotations

The following annotations are currently supported:

#### Entity annotations

* **name** - specifies the name to use in the database if it's desired to be different than what the FlatBuffer schema "table" is called.
* **transient** - this entity is skipped, no code is generated for it. Useful if you have custom FlatBuffer handling but still want to generate ObjectBox binding code for some parts of the same file.
* **uid** - used to explicitly specify UID used with this entity; used when renaming entities. See [Go documentation on schema changes](https://golang.objectbox.io/schema-changes) which applies here as well.

#### Property annotations

* **date** - tells ObjectBox the property is a timestamp, ObjectBox expects the value to be a timestamp since UNIX epoch, in milliseconds.
* **id** - specifies this property is a unique identifier of the object - used for all CRUD operations.
* **id-companion** - identifies a companion property, currently only supported on `date` properties in time-series databases.
* **index** - creates a database index. This can improve performance when querying for that property. You can specify an index type as the annotation value:
  * not specified - automatically choose the index type based on the property type (`hash` for string, `value` for others).
  * `value` - uses property values to build the index. For string, this may require more storage than a hash-based index.
  * `hash` - uses a 32-bit hash of property value to build the index. Occasional collisions may occur which should not have any performance impact in practice (with normal value distribution). Usually a better choice than `hash64`, as it requires less storage.
  * `hash64` - uses long hash of property values to build the index. Requires more storage than `hash` and thus should not be the first choice in most cases.
* **link** - declares the field as a relation ID, linking to another Entity which must be specified as a value of this annotation.
* **name** - specifies the name to use in the database if it's desired to be different than what the FlatBuffer schema "field" is called.
* **transient** - this property is skipped, no code is generated for it. Useful if you have custom FlatBuffer handling but still want to generate ObjectBox binding code for the entity.
* **uid** - used to explicitly specify UID used with this property; used when renaming properties. See [Go documentation on schema changes](https://golang.objectbox.io/schema-changes) which applies here as well.
* **unique** - set to enforce that values are unique before an entity is inserted/updated. A `put` operation will abort and return an error if the unique constraint is violated.

# License

```
Copyright (C) 2020 ObjectBox Ltd. All rights reserved.
https://objectbox.io
This file is part of ObjectBox Generator.

ObjectBox Generator is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
ObjectBox Generator is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.
You should have received a copy of the GNU General Public License
along with ObjectBox Generator.  If not, see <http://www.gnu.org/licenses/>.
```

Note: GPL only applies to the Generator itself and not to generated code.
You can regard generated code as "your code", and we impose no limitation on distributing it.
And, just to clarify: as the Generator does not include any warranty, there can be no warranty for the code it generates.       

# Do you ♥️ using ObjectBox?

We want to [hear about your project](https://docs.google.com/forms/d/e/1FAIpQLScIYiOIThcq-AnDVoCvnZOMgxO4S-fBtDSFPQfWldJnhi2c7Q/viewform)!
It will - literally - take just a minute, but help us a lot. Thank you!​ 🙏​
