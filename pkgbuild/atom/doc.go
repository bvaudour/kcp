//Package atom provides primitives to describe and get the properties
//of an atomic part of a PKGBUILD file.
//
//A PKGBUILD can contain:
//- Blank lines,
//- Comment lines,
//- Variable declarations (with distinction between string & array variables)
//  containing themselves a name and a list of values and inner comments,
//- Function declarations containing themselves a name and a body,
//- Trailing comments after variables/functions declarations.
//
//There should be only one root atom on a line of the file.
//For this reason, declarations and trailing comments are grouped in an atom
//of type group.
package atom
