function __fish_tsk_list_tasks
    # take the task name and the first line in the description and format them:
    # <name>TAB<desc>
    command tsk -l --output toml 2>/dev/null \
        | tomlq -r '
            to_entries[]
            | [ .key, (
                .value.description // ""
                | split("\n")[0]
              ) ]
            | @tsv
          '
end

complete \
    -c tsk \
    -f \
    -a '(__fish_tsk_list_tasks)'
