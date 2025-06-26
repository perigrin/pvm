---
category: typed-perl
subcategory: classes-roles
tags:
    - role-composition
    - method-conflicts
    - conflict-resolution
type_check: true
---

# Role Composition Conflicts

Role composition with method conflicts and resolution

```perl
role Drawable {
    method draw() returns Void;
    method get_bounds() returns Rectangle;
}

role Clickable {
    method on_click(Event $event) returns Void;
    method get_bounds() returns Rectangle;  # Conflict with Drawable
}

role Resizable {
    method resize(Int $width, Int $height) returns Void;
    method get_size() returns Size;
}

class Widget does Drawable, Clickable, Resizable {
    field Int $x = 0;
    field Int $y = 0;
    field Int $width = 100;
    field Int $height = 50;

    # Resolve conflict by implementing the conflicting method
    method get_bounds() returns Rectangle {
        return Rectangle->new($x, $y, $width, $height);
    }

    method draw() returns Void {
        # Implementation for drawing
    }

    method on_click(Event $event) returns Void {
        # Handle click event
    }

    method resize(Int $new_width, Int $new_height) returns Void {
        $width = $new_width;
        $height = $new_height;
    }

    method get_size() returns Size {
        return Size->new($width, $height);
    }
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 1047 characters
  Type Annotations:
    MethodReturnAnnotation: draw :: Void at 2:27
    MethodReturnAnnotation: get_bounds :: Rectangle at 3:33
    MethodReturnAnnotation: on_click :: Void at 7:43
    MethodReturnAnnotation: get_bounds :: Rectangle at 8:33
    MethodReturnAnnotation: resize :: Void at 12:52
    MethodReturnAnnotation: get_size :: Size at 13:31
    VarAnnotation: Widget :: class at 16:1
    VarAnnotation: $x :: Int at 17:5
    VarAnnotation: $y :: Int at 18:5
    VarAnnotation: $width :: Int at 19:5
    VarAnnotation: $height :: Int at 20:5
    MethodReturnAnnotation: get_bounds :: Rectangle at 23:33
    MethodReturnAnnotation: draw :: Void at 27:27
    MethodReturnAnnotation: on_click :: Void at 31:43
    MethodReturnAnnotation: resize :: Void at 35:60
    MethodReturnAnnotation: get_size :: Size at 40:31
    MethodParamAnnotation: $event :: Event at 7:1
    MethodParamAnnotation: $width :: Int at 12:1
    MethodParamAnnotation: $height :: Int at 12:1
    FieldAnnotation: $x :: Int at 17:1
    FieldAnnotation: $y :: Int at 18:1
    FieldAnnotation: $width :: Int at 19:1
    FieldAnnotation: $height :: Int at 20:1
    MethodParamAnnotation: $event :: Event at 31:1
    MethodParamAnnotation: $new_width :: Int at 35:1
    MethodParamAnnotation: $new_height :: Int at 35:1
  Root: source_file
  Tree Structure:
  source_file
    role_decl
      method_decl
      method_decl
    role_decl
      method_decl
      method_decl
    role_decl
      method_decl
      method_decl
    class_decl
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 1047 characters
  Type Annotations:
    MethodReturnAnnotation: draw :: Void at 2:27
    MethodReturnAnnotation: get_bounds :: Rectangle at 3:33
    MethodReturnAnnotation: on_click :: Void at 7:43
    MethodReturnAnnotation: get_bounds :: Rectangle at 8:33
    MethodReturnAnnotation: resize :: Void at 12:52
    MethodReturnAnnotation: get_size :: Size at 13:31
    VarAnnotation: Widget :: class at 16:1
    VarAnnotation: $x :: Int at 17:5
    VarAnnotation: $y :: Int at 18:5
    VarAnnotation: $width :: Int at 19:5
    VarAnnotation: $height :: Int at 20:5
    MethodReturnAnnotation: get_bounds :: Rectangle at 23:33
    MethodReturnAnnotation: draw :: Void at 27:27
    MethodReturnAnnotation: on_click :: Void at 31:43
    MethodReturnAnnotation: resize :: Void at 35:60
    MethodReturnAnnotation: get_size :: Size at 40:31
    MethodParamAnnotation: $event :: Event at 7:1
    MethodParamAnnotation: $width :: Int at 12:1
    MethodParamAnnotation: $height :: Int at 12:1
    FieldAnnotation: $x :: Int at 17:1
    FieldAnnotation: $y :: Int at 18:1
    FieldAnnotation: $width :: Int at 19:1
    FieldAnnotation: $height :: Int at 20:1
    MethodParamAnnotation: $event :: Event at 31:1
    MethodParamAnnotation: $new_width :: Int at 35:1
    MethodParamAnnotation: $new_height :: Int at 35:1
  Root: source_file
  Tree Structure:
  source_file
    role_decl
      method_decl
      method_decl
    role_decl
      method_decl
      method_decl
    role_decl
      method_decl
      method_decl
    class_decl
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
role Drawable {
    method draw() returns Void;
    method get_bounds() returns Rectangle;
}

role Clickable {
    method on_click(Event $event) returns Void;
    method get_bounds() returns Rectangle;  # Conflict with Drawable
}

role Resizable {
method resize(Int $width, $height) returns Void;
    method get_size() returns Size;
}

class Widget does Drawable, Clickable, Resizable {
field $x = 0;
field $y = 0;
field $width = 100;
field $height = 50;

    # Resolve conflict by implementing the conflicting method
    method get_bounds() returns Rectangle {
return $y, $width, $height);
    }

    method draw() returns Void {
        # Implementation for drawing
    }

    method on_click(Event $event) returns Void {
        # Handle click event
    }

method resize(Int $new_width, $new_height) returns Void {
        $width = $new_width;
        $height = $new_height;
    }

    method get_size() returns Size {
return $height);
    }
}
```

## Typed Perl Output

```perl
role Drawable {
    method draw() returns Void;
    method get_bounds() returns Rectangle;
}

role Clickable {
    method on_click(Event $event) returns Void;
    method get_bounds() returns Rectangle;  # Conflict with Drawable
}

role Resizable {
    method resize(Int $width, Int $height) returns Void;
    method get_size() returns Size;
}

class Widget does Drawable, Clickable, Resizable {
    field Int $x = 0;
    field Int $y = 0;
    field Int $width = 100;
    field Int $height = 50;

    # Resolve conflict by implementing the conflicting method
    method get_bounds() returns Rectangle {
        return Rectangle->new($x, $y, $width, $height);
    }

    method draw() returns Void {
        # Implementation for drawing
    }

    method on_click(Event $event) returns Void {
        # Handle click event
    }

    method resize(Int $new_width, Int $new_height) returns Void {
        $width = $new_width;
        $height = $new_height;
    }

    method get_size() returns Size {
        return Size->new($width, $height);
    }
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
