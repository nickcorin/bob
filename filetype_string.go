// Code generated by "stringer -type FileType -trimprefix FileType"; DO NOT EDIT.

package bob

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[FileTypeUnknown-0]
	_ = x[FileTypeDockerCompose-1]
	_ = x[FileTypeDockerfile-2]
	_ = x[FileTypeMakefile-3]
	_ = x[FileTypeServiceDiscovery-4]
	_ = x[fileTypeSentinel-5]
}

const _FileType_name = "UnknownDockerComposeDockerfileMakefileServiceDiscoveryfileTypeSentinel"

var _FileType_index = [...]uint8{0, 7, 20, 30, 38, 54, 70}

func (i FileType) String() string {
	if i < 0 || i >= FileType(len(_FileType_index)-1) {
		return "FileType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _FileType_name[_FileType_index[i]:_FileType_index[i+1]]
}
