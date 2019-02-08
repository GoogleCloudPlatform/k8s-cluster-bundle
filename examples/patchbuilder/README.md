# patch template builder example

Starting with a ComponentBuilder and a PatchTemplateBuilder, we execute

```
$ bundlectl build --input-file builder.yaml --options-file build-options.yaml
```

which results in a Component that includes a PatchTemplate. This we can apply
using `bundlectl patch`. In this case we have a schema in the PatchTemplate that
has a default value, so no patch-time options are needed:

```
$ bundlectl build --input-file builder.yaml --options-file build-options.yaml | \
        bundlectl patch
```

This results in a Component patched according to the PatchTemplate.
