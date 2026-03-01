The talk remains generally correct, however there have been a few changes to the linker since 2023 that concern it:

* A new method, `reflect.Value.Methods` has been introduced that has the same effect as `reflect.Value.Method`.
* Calls to `reflect.Value.Method`, `reflect.Value.MethodByName` and `reflect.Value.Methods` can now be inlined, the functions they are inlined into will be flagged in a special way that partially disables DCE like they `reflect` methods did.
* Passing a constant string to `reflect.Value.MethodByName` no longer partially disables DCE, instead it works similarly to making an interface reachable (all methods with the name passed to `MethodByName` will be retained, but other public methods can still be pruned).

