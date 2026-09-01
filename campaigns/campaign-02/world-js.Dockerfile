# world-js.Dockerfile — the JS resolution layer over the campaign
# world: node's upward module lookup from the throwaway run
# directories (/tmp/check-*) reaches /tmp/node_modules — a permanent
# symlink to the world's libraries at /run/node_modules. Without it,
# bare imports resolve only under /run/task and every js task needs a
# library-level fallback.
#
# Campaign rule: the task environment carries its own stack; the
# machine neither knows nor decides this.
#
# Build: docker build -f campaigns/campaign-02/world-js.Dockerfile \
#                    -t punchtape-world:c02js /tmp/c02ctx
FROM punchtape-world:c02

RUN ln -s /run/node_modules /tmp/node_modules
