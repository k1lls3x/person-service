CREATE TABLE persons (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    surname VARCHAR(100) NOT NULL,
    patronymic VARCHAR(100),
    age INT,
    gender VARCHAR(20),
    nationality VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_persons_main_search ON persons(surname, name);
CREATE INDEX idx_persons_created_at_desc ON persons(created_at DESC);  -- для сортировки новых записей

CREATE INDEX idx_persons_age_partial ON persons(age) WHERE age IS NOT NULL;
CREATE INDEX idx_persons_surname ON persons(surname);
