---
authors: Alex McGrath (alex.mcgrath@goteleport.com)
state: draft
---

# RFD 57 - Automatic user and sudoers provisioning

## What

Automatically create non-existing users and optionally add them to
`sudoers` on Teleport nodes. Users will be removed after all sessions
have logged out.

## Why

Currently, when logging into an SSH node, the user must be
pre-created. Adding automatic user and `sudoer` provisioning would
make it so that any Teleport user would be able to login and have the
account created automatically without manual intervention.

## Details

The following are required for this feature:

- Ability to automatically provision a Linux user if it's not present
  on the node.
- Ability to add the provisioned user to existing Linux groups defined
  in the user traits/role.
- Ability to add the provisioned user to sudoers.
- Clean up the provisioned user / sudoers changes upon logout (being
  careful not to remove pre-existing users).

### Config/Role Changes

Several new fields will need to be added to to the role `options` and
`allow` sections:

```yaml
kind: role
version: v5
metadata:
  name: example
spec:
  options:
    # Controls whether this role supports auto provisioning of users.
    auto_create_user: true
    # Controls whether this role allows adding auto provisioned users to sudoers.
    add_to_sudoers: true
  allow:
    # New field listing Linux groups to assign a provisioned user to.
    # Should support user and identity provider traits like other fields (e.g. "logins")
    groups: [ubuntu, "{{internal.groups}}", "{{external.xxx}}"]
```

An individual `ssh_service` can be configured disable auto user
creation with the below config:

```yaml
ssh_service:
    # when disabled, takes precedence over the role setting
    disable_auto_create_user: true
```

### User creation

In order to create users `useradd` will be executed from teleport
after a user has tried to access a Teleport SSH node.

#### User Groups

When a user is created they will be added to the specified groups from
the role. In addition the user will be added to a special
`teleport-system` group which can be used to indicate that the user
was created by teleport and that its safe for it to be deleted. The
groups will be created via `groupadd` at startup if they do not already
exist and users will be added to groups via `usermod -aG <list of groups> <username>`

#### Valid user/group names

The set of valid names that are valid on Linux varies between distros
and are generally more restrictive than the allowed usernames in
Teleport. This will require that names containing invalid characters
have those characters removed/replaced. Information on the valid
characters between Linux distros is available [here](https://systemd.io/USER_NAMES/).
The common core of valid characters is `^[a-z][a-z0-9-]{0,30}$`.

#### Adding and removing users from sudoers

A user can be added to `sudoers` by creating files for each user in
`/etc/sudoers.d/`.

Each file will contain the following contents, allowing the specified
user to execute root commands without a password:
```
${USERNAME} ALL=(ALL) ALL
```

##### Other considerations

`sudo` isn't always available and [doas](https://man.openbsd.org/doas.conf.5)
can be used similarly to `sudo`.  Supporting systems using `doas`
would required modifying `/etc/doas.conf` to include `permit` line:

```
permit nopass ${USERNAME} as root
```

On nodes with only `su` available a PAM module exists that
allows for users in a particular group -- such as `sudoers` or
`wheel`, may execute `su` commands.

### User deletion

After all of a users sessions are logged out the created user and any
`sudoers` files that were created for that user will be deleted if
that user is also a member of the `teleport-system` group.

Users can not be deleted while they have running processes so each
time a session ends, an attempt to delete the user can happen, if it
succeeds the sudoers file can also be removed.

If it does not succeed a cleanup process will run that will check for
users with no active sessions that also have active processes, kill
the processes and remove their user and sudoer files.
