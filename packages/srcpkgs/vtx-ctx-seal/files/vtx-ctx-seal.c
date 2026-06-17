// vtx-ctx-seal — VertexOS FUSE context-seal.
//
// A passthrough filesystem that mirrors a source directory (the real $HOME) to a
// mountpoint, but REDACTS a denylist of credential paths: any sealed path returns
// ENOENT (it appears not to exist) and is filtered out of directory listings.
//
// Why FUSE and not a bind-mount/bwrap seal: VertexOS sets user.max_user_namespaces=0
// (white-team V7 LPE defense), so an unprivileged agent CANNOT create a mount
// namespace. FUSE via the setuid fusermount3 helper is the one redaction mechanism
// that still works unprivileged. vtx-sandbox mounts this at a private dir and runs
// the agent with HOME pointed at it; the AppArmor profile is the hard backstop for
// absolute-path access.
//
//   VTX_SEAL_SRC=/home/vertex vtx-ctx-seal /run/user/1000/vtx-home
//   VTX_SEAL_PATHS=".ssh:.aws:..."   # optional override of the denylist
//
// Build: gcc -O2 -Wall -o vtx-ctx-seal vtx-ctx-seal.c $(pkg-config fuse3 --cflags --libs)

#define FUSE_USE_VERSION 31
#define _GNU_SOURCE

#include <fuse.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <fcntl.h>
#include <errno.h>
#include <dirent.h>
#include <sys/stat.h>
#include <sys/statvfs.h>

static char src_root[4096];

#define MAX_SEAL 64
static char *seal_list[MAX_SEAL];
static int   seal_n = 0;

static const char *DEFAULT_SEAL =
    ".ssh:.aws:.gnupg:.config/gh:.config/gcloud:.kube:.azure:.docker/config.json:"
    ".config/git/credentials:.netrc:.password-store:.config/op:.gh:"
    ".mozilla:.config/google-chrome:.config/chromium:.config/BraveSoftware";

static void load_seal(void) {
    const char *env = getenv("VTX_SEAL_PATHS");
    char *s = strdup(env && *env ? env : DEFAULT_SEAL);
    char *tok = strtok(s, ":");
    while (tok && seal_n < MAX_SEAL) {
        if (*tok) seal_list[seal_n++] = strdup(tok);
        tok = strtok(NULL, ":");
    }
    free(s);
}

// path is a FUSE path ("/", "/.ssh", "/.ssh/id_rsa", "/projects/.env").
static int is_sealed(const char *path) {
    if (path[0] != '/') return 0;
    const char *rel = path + 1;
    if (*rel == '\0') return 0;                 // root is never sealed

    // Seal any .env file regardless of directory (matches the AppArmor profile).
    const char *base = strrchr(rel, '/');
    base = base ? base + 1 : rel;
    size_t blen = strlen(base);
    if (strcmp(base, ".env") == 0) return 1;
    if (blen >= 4 && strcmp(base + blen - 4, ".env") == 0) return 1;

    // Denylist prefixes: exact match or a child of the sealed dir.
    for (int i = 0; i < seal_n; i++) {
        size_t el = strlen(seal_list[i]);
        if (strncmp(rel, seal_list[i], el) == 0 && (rel[el] == '\0' || rel[el] == '/'))
            return 1;
    }
    return 0;
}

static void real_path(char *out, size_t n, const char *path) {
    snprintf(out, n, "%s%s", src_root, path);
}

static int vx_getattr(const char *path, struct stat *st, struct fuse_file_info *fi) {
    (void)fi;
    if (is_sealed(path)) return -ENOENT;
    char rp[4096]; real_path(rp, sizeof rp, path);
    return lstat(rp, st) == -1 ? -errno : 0;
}

static int vx_access(const char *path, int mask) {
    if (is_sealed(path)) return -ENOENT;
    char rp[4096]; real_path(rp, sizeof rp, path);
    return access(rp, mask) == -1 ? -errno : 0;
}

static int vx_readlink(const char *path, char *buf, size_t size) {
    if (is_sealed(path)) return -ENOENT;
    char rp[4096]; real_path(rp, sizeof rp, path);
    int r = readlink(rp, buf, size - 1);
    if (r == -1) return -errno;
    buf[r] = '\0';
    return 0;
}

static int vx_readdir(const char *path, void *buf, fuse_fill_dir_t filler,
                      off_t offset, struct fuse_file_info *fi,
                      enum fuse_readdir_flags flags) {
    (void)offset; (void)fi; (void)flags;
    if (is_sealed(path)) return -ENOENT;
    char rp[4096]; real_path(rp, sizeof rp, path);
    DIR *dp = opendir(rp);
    if (!dp) return -errno;
    struct dirent *de;
    while ((de = readdir(dp)) != NULL) {
        char child[4096];
        if (strcmp(path, "/") == 0) snprintf(child, sizeof child, "/%s", de->d_name);
        else                        snprintf(child, sizeof child, "%s/%s", path, de->d_name);
        // Always show . and .. ; hide sealed children entirely.
        if (strcmp(de->d_name, ".") != 0 && strcmp(de->d_name, "..") != 0 && is_sealed(child))
            continue;
        struct stat st; memset(&st, 0, sizeof st);
        st.st_ino = de->d_ino; st.st_mode = de->d_type << 12;
        if (filler(buf, de->d_name, &st, 0, 0)) break;
    }
    closedir(dp);
    return 0;
}

static int vx_open(const char *path, struct fuse_file_info *fi) {
    if (is_sealed(path)) return -ENOENT;
    char rp[4096]; real_path(rp, sizeof rp, path);
    int fd = open(rp, fi->flags);
    if (fd == -1) return -errno;
    fi->fh = fd;
    return 0;
}

static int vx_create(const char *path, mode_t mode, struct fuse_file_info *fi) {
    if (is_sealed(path)) return -EACCES;            // never let the agent create a sealed path
    char rp[4096]; real_path(rp, sizeof rp, path);
    int fd = open(rp, fi->flags, mode);
    if (fd == -1) return -errno;
    fi->fh = fd;
    return 0;
}

static int vx_read(const char *path, char *buf, size_t size, off_t off,
                   struct fuse_file_info *fi) {
    (void)path;
    int r = pread(fi->fh, buf, size, off);
    return r == -1 ? -errno : r;
}

static int vx_write(const char *path, const char *buf, size_t size, off_t off,
                    struct fuse_file_info *fi) {
    (void)path;
    int r = pwrite(fi->fh, buf, size, off);
    return r == -1 ? -errno : r;
}

static int vx_release(const char *path, struct fuse_file_info *fi) {
    (void)path;
    if (fi->fh) close(fi->fh);
    return 0;
}

static int vx_fsync(const char *path, int ds, struct fuse_file_info *fi) {
    (void)path;
    int r = ds ? fdatasync(fi->fh) : fsync(fi->fh);
    return r == -1 ? -errno : 0;
}

static int vx_truncate(const char *path, off_t size, struct fuse_file_info *fi) {
    if (is_sealed(path)) return -ENOENT;
    if (fi) return ftruncate(fi->fh, size) == -1 ? -errno : 0;
    char rp[4096]; real_path(rp, sizeof rp, path);
    return truncate(rp, size) == -1 ? -errno : 0;
}

static int vx_mkdir(const char *path, mode_t mode) {
    if (is_sealed(path)) return -EACCES;
    char rp[4096]; real_path(rp, sizeof rp, path);
    return mkdir(rp, mode) == -1 ? -errno : 0;
}

static int vx_unlink(const char *path) {
    if (is_sealed(path)) return -ENOENT;
    char rp[4096]; real_path(rp, sizeof rp, path);
    return unlink(rp) == -1 ? -errno : 0;
}

static int vx_rmdir(const char *path) {
    if (is_sealed(path)) return -ENOENT;
    char rp[4096]; real_path(rp, sizeof rp, path);
    return rmdir(rp) == -1 ? -errno : 0;
}

static int vx_symlink(const char *from, const char *to) {
    if (is_sealed(to)) return -EACCES;
    char rp[4096]; real_path(rp, sizeof rp, to);
    return symlink(from, rp) == -1 ? -errno : 0;
}

static int vx_rename(const char *from, const char *to, unsigned int flags) {
    if (flags) return -EINVAL;
    if (is_sealed(from) || is_sealed(to)) return -EACCES;
    char rf[4096], rt[4096]; real_path(rf, sizeof rf, from); real_path(rt, sizeof rt, to);
    return rename(rf, rt) == -1 ? -errno : 0;
}

static int vx_link(const char *from, const char *to) {
    if (is_sealed(from) || is_sealed(to)) return -EACCES;
    char rf[4096], rt[4096]; real_path(rf, sizeof rf, from); real_path(rt, sizeof rt, to);
    return link(rf, rt) == -1 ? -errno : 0;
}

static int vx_chmod(const char *path, mode_t mode, struct fuse_file_info *fi) {
    (void)fi;
    if (is_sealed(path)) return -ENOENT;
    char rp[4096]; real_path(rp, sizeof rp, path);
    return chmod(rp, mode) == -1 ? -errno : 0;
}

static int vx_chown(const char *path, uid_t uid, gid_t gid, struct fuse_file_info *fi) {
    (void)fi;
    if (is_sealed(path)) return -ENOENT;
    char rp[4096]; real_path(rp, sizeof rp, path);
    return lchown(rp, uid, gid) == -1 ? -errno : 0;
}

static int vx_utimens(const char *path, const struct timespec ts[2], struct fuse_file_info *fi) {
    (void)fi;
    if (is_sealed(path)) return -ENOENT;
    char rp[4096]; real_path(rp, sizeof rp, path);
    return utimensat(0, rp, ts, AT_SYMLINK_NOFOLLOW) == -1 ? -errno : 0;
}

static int vx_statfs(const char *path, struct statvfs *st) {
    char rp[4096]; real_path(rp, sizeof rp, path);
    return statvfs(rp, st) == -1 ? -errno : 0;
}

static const struct fuse_operations vx_oper = {
    .getattr  = vx_getattr,
    .access   = vx_access,
    .readlink = vx_readlink,
    .readdir  = vx_readdir,
    .open     = vx_open,
    .create   = vx_create,
    .read     = vx_read,
    .write    = vx_write,
    .release  = vx_release,
    .fsync    = vx_fsync,
    .truncate = vx_truncate,
    .mkdir    = vx_mkdir,
    .unlink   = vx_unlink,
    .rmdir    = vx_rmdir,
    .symlink  = vx_symlink,
    .rename   = vx_rename,
    .link     = vx_link,
    .chmod    = vx_chmod,
    .chown    = vx_chown,
    .utimens  = vx_utimens,
    .statfs   = vx_statfs,
};

int main(int argc, char *argv[]) {
    const char *s = getenv("VTX_SEAL_SRC");
    if (!s || !*s) {
        fprintf(stderr, "vtx-ctx-seal: set VTX_SEAL_SRC to the real source dir (e.g. $HOME)\n");
        return 2;
    }
    // Strip a trailing slash so real_path concatenation stays clean.
    snprintf(src_root, sizeof src_root, "%s", s);
    size_t l = strlen(src_root);
    if (l > 1 && src_root[l - 1] == '/') src_root[l - 1] = '\0';

    struct stat st;
    if (lstat(src_root, &st) == -1 || !S_ISDIR(st.st_mode)) {
        fprintf(stderr, "vtx-ctx-seal: source '%s' is not a directory\n", src_root);
        return 2;
    }
    load_seal();
    return fuse_main(argc, argv, &vx_oper, NULL);
}
