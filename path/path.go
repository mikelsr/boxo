// Package path contains utilities to work with ipfs paths.
package path

import (
	"fmt"
	gopath "path"
	"strings"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	IPFSNamespace = "ipfs"
	IPNSNamespace = "ipns"
	IPLDNamespace = "ipld"
)

// Path is a generic path. Paths must be prefixed with a valid prefix.
type Path interface {
	// String returns the path as a string.
	String() string

	// Namespace returns the first component of the path. For example, the namespace
	// of "/ipfs/bafy" is "ipfs".
	Namespace() string

	// Mutable returns false if the data pointed to by this path is guaranteed to not
	// change. Note that resolved mutable paths can be immutable.
	Mutable() bool

	// Root returns the [cid.Cid] of the root object of the path. Root can return
	// [cid.Undef] for Mutable IPNS paths that use [DNSLink].
	//
	// [DNSLink]: https://dnslink.dev/
	Root() cid.Cid

	// Segments returns the different elements ofd a path, which are delimited
	// by a forward slash ("/"). The leading slash must be ignored, that is, no
	// segment should be empty.
	Segments() []string
}

// ResolvedPath is a [Path] which was resolved to the last resolvable node.
type ResolvedPath interface {
	Path

	// Cid returns the [cid.Cid] of the node referenced by the path.
	Cid() cid.Cid

	// Remainder returns the unresolved parts of the path.
	Remainder() string
}

// ImmutablePath is a [Path] which is guaranteed to return "false" to [Mutable].
type ImmutablePath struct {
	Path
}

func NewImmutablePath(p Path) (ImmutablePath, error) {
	if p.Mutable() {
		return ImmutablePath{}, fmt.Errorf("path was expected to be immutable: %s", p.String())
	}

	return ImmutablePath{p}, nil
}

type path struct {
	str       string
	root      cid.Cid
	namespace string
}

func (p *path) String() string {
	return p.str
}

func (p *path) Namespace() string {
	return p.namespace
}

func (p *path) Mutable() bool {
	return p.namespace == IPNSNamespace
}

func (p *path) Root() cid.Cid {
	return p.root
}

func (p *path) Segments() []string {
	// Trim slashes from beginning and end, such that we do not return empty segments.
	str := strings.TrimSuffix(p.str, "/")
	str = strings.TrimPrefix(str, "/")

	return strings.Split(str, "/")
}

type resolvedPath struct {
	path
	cid       cid.Cid
	remainder string
}

func (p *resolvedPath) Cid() cid.Cid {
	return p.cid
}

func (p *resolvedPath) Remainder() string {
	return p.remainder
}

// NewIPFSPath returns a new "/ipfs" path with the provided CID.
func NewIPFSPath(cid cid.Cid) ResolvedPath {
	return &resolvedPath{
		path: path{
			str:       fmt.Sprintf("/%s/%s", IPFSNamespace, cid.String()),
			root:      cid,
			namespace: IPFSNamespace,
		},
		cid:       cid,
		remainder: "",
	}
}

// NewIPLDPath returns a new "/ipld" path with the provided CID.
func NewIPLDPath(cid cid.Cid) ResolvedPath {
	return &resolvedPath{
		path: path{
			str:       fmt.Sprintf("/%s/%s", IPLDNamespace, cid.String()),
			root:      cid,
			namespace: IPLDNamespace,
		},
		cid:       cid,
		remainder: "",
	}
}

// NewIPNSPath returns a new "/ipns" path with the provided CID.
func NewIPNSPath(cid cid.Cid) Path {
	return &path{
		str:       fmt.Sprintf("/%s/%s", IPNSNamespace, cid.String()),
		root:      cid,
		namespace: IPNSNamespace,
	}
}

func NewDNSLinkPath(domain string) Path {
	return &path{
		str:       fmt.Sprintf("/%s/%s", IPNSNamespace, domain),
		root:      cid.Undef,
		namespace: IPNSNamespace,
	}
}

// NewPath returns a well-formed [Path]. The returned path will always be prefixed
// with a valid namespace (/ipfs, /ipld, or /ipns). The prefix will be added if not
// present in the given string. The rules are:
//
//  1. If the path has a single component (no slashes) ans it is a valid CID,
//     an /ipfs path is returned. If the CID is encoded with the Libp2pKey codec,
//     then a /ipns path is returned.
//  2. If the path has a valid CID root but does not have a namespace, the /ipfs
//     namespace is automatically added.
//
// This function returns an error when the given string is not a valid path.
func NewPath(str string) (Path, error) {
	cleaned := gopath.Clean(str)
	components := strings.Split(cleaned, "/")

	if strings.HasSuffix(str, "/") {
		// Do not forget to store the trailing slash!
		cleaned += "/"
	}

	// If there's only one component, check if it's a CID, or Peer ID.
	if len(components) == 1 {
		c, err := cid.Decode(components[0])
		if err == nil {
			if c.Prefix().GetCodec() == cid.Libp2pKey {
				return NewIPNSPath(c), nil
			} else {
				return NewIPFSPath(c), nil
			}
		}
	}

	// If the path doesn't begin with a "/", we expect it to start with a CID and
	// be an IPFS Path.
	if components[0] != "" {
		root, err := cid.Decode(components[0])
		if err != nil {
			return nil, &ErrInvalidPath{error: err, path: str}
		}

		return &path{
			str:       cleaned,
			root:      root,
			namespace: IPFSNamespace,
		}, nil
	}

	if len(components) < 3 {
		return nil, &ErrInvalidPath{error: fmt.Errorf("not enough path components"), path: str}
	}

	switch components[1] {
	case IPFSNamespace, IPLDNamespace:
		if components[2] == "" {
			return nil, &ErrInvalidPath{error: fmt.Errorf("not enough path components"), path: str}
		}

		root, err := cid.Decode(components[2])
		if err != nil {
			return nil, &ErrInvalidPath{error: fmt.Errorf("invalid CID: %w", err), path: str}
		}

		return &path{
			str:       cleaned,
			root:      root,
			namespace: components[1],
		}, nil
	case IPNSNamespace:
		if components[2] == "" {
			return nil, &ErrInvalidPath{error: fmt.Errorf("not enough path components"), path: str}
		}

		var root cid.Cid
		pid, err := peer.Decode(components[2])
		if err != nil {
			// DNSLink.
			root = cid.Undef
		} else {
			root = peer.ToCid(pid)
		}

		return &path{
			str:       cleaned,
			root:      root,
			namespace: IPFSNamespace,
		}, nil
	default:
		return nil, &ErrInvalidPath{error: fmt.Errorf("unknown namespace %q", components[1]), path: str}
	}
}

// NewPathFromSegments creates a new [Path] from the provided segments. This
// function simply calls [NewPath] internally with the segments concatenated
// using a forward slash "/" as separator.
func NewPathFromSegments(segments ...string) (Path, error) {
	if len(segments) > 1 {
		if segments[0] == "" {
			segments = segments[1:]
		}
	}

	return NewPath("/" + strings.Join(segments, "/"))
}

// SplitImmutablePath cleans up and splits the given path. It extracts the first
// component, which must be a CID, and returns it separately.
func SplitImmutablePath(fpath Path) (cid.Cid, []string, error) {
	// TODO: probably rewrite this and use the .Namespace and .Root.

	parts := fpath.Segments()
	if parts[0] == IPFSNamespace || parts[0] == IPLDNamespace {
		parts = parts[1:]
	}

	// if nothing, bail.
	if len(parts) == 0 {
		return cid.Undef, nil, &ErrInvalidPath{error: fmt.Errorf("empty"), path: fpath.String()}
	}

	c, err := cid.Decode(parts[0])
	// first element in the path is a cid
	if err != nil {
		return cid.Undef, nil, &ErrInvalidPath{error: fmt.Errorf("invalid CID: %w", err), path: fpath.String()}
	}

	return c, parts[1:], nil
}

func Join(p Path, segments ...string) (Path, error) {
	s := p.Segments()
	s = append(s, segments...)
	return NewPathFromSegments(s...)
}

func NewResolvedPath(p Path, cid cid.Cid, remainder string) ResolvedPath {
	return &resolvedPath{
		path: path{
			str:       fmt.Sprintf("%s/%s", p.String(), remainder),
			root:      p.Root(),
			namespace: p.Namespace(),
		},
		cid:       cid,
		remainder: remainder,
	}
}
