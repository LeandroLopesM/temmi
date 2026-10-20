(define firstTime #t)
(define treeRounds 0)

(passage "start"
    ; This doesn't run when the function is loaded on engine boot
    (if firstTime
        (begin
            (text "You wake up on nuzzled against a tree. Getting up, you get a very strong urge to go around it.")
            (set! firstTime #f)
        )
    ; else
        (set-text! "You walked around the tree " treeRounds " times."))

    (choice "behind-tree" "Walk around the tree")
)

(passage "behind-tree"
    (text "You are behind the tree")
    (choice "start" "Continue..." '(set! treeRounds (+ treeRounds 1)))
)