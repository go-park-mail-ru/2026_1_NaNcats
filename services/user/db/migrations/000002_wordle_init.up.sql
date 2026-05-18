CREATE TABLE "wordle_word" (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    word TEXT NOT NULL UNIQUE
        CHECK (char_length(word) = 5 AND word = LOWER(word)),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE "wordle_daily" (
    word_of_day DATE PRIMARY KEY,
    word_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    CONSTRAINT fk_wordle_daily_wordle_word
        FOREIGN KEY (word_id)
        REFERENCES "wordle_word"(id)
        ON DELETE RESTRICT
);

CREATE TABLE "wordle_game" (
    user_id BIGINT NOT NULL,
    game_date DATE NOT NULL,
    PRIMARY KEY(user_id, game_date),
    
    solved BOOL NOT NULL DEFAULT FALSE,
    attempt INT NOT NULL DEFAULT 0
        CHECK (attempt >= 0 AND attempt <= 6),
        
    finished_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    
    CONSTRAINT fk_wordle_game_user
        FOREIGN KEY (user_id)
        REFERENCES "user"(id)
        ON DELETE CASCADE,
    
    CONSTRAINT fk_wordle_game_wordle_daily
        FOREIGN KEY (game_date)
        REFERENCES "wordle_daily"(word_of_day)
        ON DELETE RESTRICT
);

CREATE TABLE "wordle_guess" (
    user_id BIGINT NOT NULL,
    guess_date DATE NOT NULL,
    attempt_num INT NOT NULL
        CHECK (attempt_num >= 1 AND attempt_num <= 6),
    PRIMARY KEY(user_id, guess_date, attempt_num),
        
    word TEXT NOT NULL
        CHECK (char_length(word) = 5 AND word = LOWER(word)),
        
    idempotency_key TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    
    CONSTRAINT fk_wordle_guess_wordle_game
        FOREIGN KEY (user_id, guess_date)
        REFERENCES "wordle_game"(user_id, game_date)
        ON DELETE CASCADE 
);

CREATE TABLE "wordle_streak" (
    user_id BIGINT PRIMARY KEY,
    current_streak INT NOT NULL DEFAULT 0,
    last_played DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    
    CONSTRAINT fk_wordle_streak_client_profile
        FOREIGN KEY (user_id)
        REFERENCES "user"(id)
        ON DELETE CASCADE
);
