(set-opt! responsive #f)
(set-opt! lineWidth 100)
(define pain 10.0)
(define integrity 1.0)

(passage "start"
    (text "You wake up in a cold sweat on a hardwood floor. Your mouth feels weird, swolen. Aching?")
    (text "\nYou probe your mouth with your tongue only to feel blood pour out of your mouth under the disturbance.")
    (text "\nYou turn stomach down to the floor, spending a few minutes hacking and coughing up blood from your throat.")
    (choice "coself" "Collect yourself")
    (choice "around" "Scan your surroundings")
)

(passage "coself"
    (on-load '(define pain (+ pain 10.0)))
    (on-load '(define integrity (+ integrity 1.0)))

    (text "Your back arches against the moonlight, you can feel your bones shifting beneath your skin.")
    (text "\nIts night... right? Yes... you think it is. There is a throbbing pain in the back of your head which seems to reverberate throughout your entire body.")
    (text "\nThe pains slowly shifts. It doesn't lessen or disappear, you just retain your focus on something else. You can't remember your name.")
    (choice "remember" "Try to remember")
    (choice "around" "Scan your surroundings")
)

(passage "remember"
    (text "Who are you? Why are you here? What happened to your body?")
    (text "\nYou wrack your brain for any answers only to be met with the same ache that spreads through your muscles.")
    (text "\nEverytime you try to remember anything, even your name, only one word appears: " (ul "Anar") ".\n\n")

    (on-load '(define integrity (+ integrity 3.0)))
    (text (dark-gray "You pondered the ever-shifting self! ") (green "+3 Integrity") "\n")
    (on-load '(define pain (+ pain 20.0)))
    (text (dark-gray "You paid the toll for pondering. ") (red "+20 Pain"))

    (choice "around" "Scan your surroundings")
)

(passage "around"
    (text "You are laying on a cheap lynoleum floor, beside a tall bed amidst a sea if books, newspapers and other paraphernalia. The room is barely visible, the only visibility is afforded by the thin streams of light peering through a curtain on the other end of the room.")

    (choice "sdup" "Stand up")
    (choice "ubed" "Look under the bed")
)

(passage "ubed"
    (text "Since you are already on the floor, you reach out a bony, pale finger and lift the dangling edge of a blanket.")
    (text "\n\nSurprising no one, you can't see anything. But you sure do feel like something saw you.")

    (on-load (define integrity (- 3.0 integrity)))
    (text (dark-gray "The void stared back! ") (red "-3 Integrity"))

    (choice "sdup" "Stand up")
)