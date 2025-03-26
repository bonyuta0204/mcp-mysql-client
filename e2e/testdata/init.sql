
CREATE DATABASE IF NOT EXISTS testdb;

USE testdb;

-- Drop tables if they exist
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS products;

CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert sample data
INSERT INTO users (username, email) VALUES
    ('user1', 'user1@example.com'),
    ('user2', 'user2@example.com'),
    ('user3', 'user3@example.com');

-- Create products table
CREATE TABLE products (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert sample data
INSERT INTO products (name, price, description) VALUES
    ('Product A', 19.99, 'Description for Product A'),
    ('Product B', 29.99, 'Description for Product B'),
    ('Product C', 39.99, 'Description for Product C');

-- Create a table for testing special data types
USE testdb;

-- Drop the table if it exists
DROP TABLE IF EXISTS data_types;

-- Create a table with various data types
CREATE TABLE data_types (
    id INT AUTO_INCREMENT PRIMARY KEY,
    nullable_column VARCHAR(100) NULL,
    integer_column INT NOT NULL,
    float_column FLOAT NOT NULL,
    decimal_column DECIMAL(10, 2) NOT NULL
);

-- Insert sample data including NULL values, integers, and floats
INSERT INTO data_types (nullable_column, integer_column, float_column, decimal_column) VALUES
    (NULL, 42, 3.14, 99.99),
    ('Not Null', 0, 0.0, 0.00),
    ('String Value', -100, -1.5, -199.99),
    (NULL, 9999, 123.456, 1000.01);

-- Create a second database for testing database switching
CREATE DATABASE IF NOT EXISTS seconddb;

USE seconddb;

-- Drop tables if they exist
DROP TABLE IF EXISTS items;

-- Create a table in the second database
CREATE TABLE items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);

-- Insert sample data
INSERT INTO items (name) VALUES
    ('Item 1'),
    ('Item 2'),
    ('Item 3');
