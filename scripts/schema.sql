-- Initialize the database.

CREATE TABLE user (
  id INT PRIMARY KEY AUTO_INCREMENT,
  username VARCHAR(20) UNIQUE NOT NULL,
  password VARCHAR(120) DEFAULT '123456',
  created TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE category (
  id INT PRIMARY KEY AUTO_INCREMENT,
  category VARCHAR(15) NOT NULL,
  created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  user_id INT NOT NULL
);

CREATE TABLE bookmark (
  id INT PRIMARY KEY AUTO_INCREMENT,
  bookmark VARCHAR(40) NOT NULL,
  url TEXT NOT NULL,
  created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  user_id INT NOT NULL,
  category_id INT DEFAULT 0
);

CREATE TABLE seq (
  user_id INT NOT NULL,
  bookmark_id INT NOT NULL,
  seq INT NOT NULL
);

DELIMITER ;;
CREATE TRIGGER add_user AFTER INSERT ON user
FOR EACH ROW BEGIN
    INSERT INTO bookmark
      (user_id, bookmark, url)
    VALUES
      (new.id, 'Google', 'https://www.google.com');
END;;

CREATE TRIGGER add_seq AFTER INSERT ON bookmark
FOR EACH ROW BEGIN
    SET @seq := (SELECT IFNULL(MAX(seq)+1, 1) FROM seq WHERE user_id = new.user_id);
    INSERT INTO seq
      (user_id, bookmark_id, seq)
    VALUES
      (new.user_id, new.id, @seq);
END;;

CREATE TRIGGER reorder AFTER DELETE ON bookmark
FOR EACH ROW BEGIN
    SET @seq := (SELECT seq FROM seq WHERE user_id = old.user_id AND bookmark_id = old.id);
    DELETE FROM seq
    WHERE user_id = old.user_id AND seq = @seq;
    UPDATE seq SET seq = seq-1
    WHERE user_id = old.user_id AND seq > @seq;
END;;
DELIMITER ;

INSERT INTO user (id, username)
  VALUES (0, 'guest');
