function __fish_tsk_list_tasks
    # take the task name and the first line in the description and format them:
    # <name>TAB<desc>
    command /Users/nate/code/tsk/bin/tsk -l --output json 2>/dev/null \
        | jq -r 'to_entries[] | [.key, .value.Desc] | @tsv'
end

complete \
    -c tsk \
    -f \
    -a '(__fish_tsk_list_tasks)'
