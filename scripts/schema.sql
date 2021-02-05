-- Initialize the database.

CREATE TABLE user (
  id INT PRIMARY KEY AUTO_INCREMENT,
  username VARCHAR(20) UNIQUE NOT NULL,
  password VARCHAR(120) DEFAULT '123456',
  uid VARCHAR(24) UNIQUE NOT NULL
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

CREATE VIEW bookmarks AS
  SELECT bookmark.user_id, bookmark.id bookmark_id, bookmark, url, bookmark.category_id, category, seq
  FROM bookmark LEFT JOIN category ON bookmark.category_id = category.id
  LEFT JOIN seq ON bookmark.user_id = seq.user_id AND bookmark.id = seq.bookmark_id
  ORDER BY seq;

CREATE VIEW categories AS
  SELECT category.id, category.user_id, category, COUNT(bookmark) count
  FROM category LEFT JOIN bookmark ON category.id = category_id
  GROUP BY category ORDER BY category;

DELIMITER ;;
CREATE PROCEDURE delete_category (cid INT)
BEGIN
  DECLARE EXIT HANDLER FOR SQLEXCEPTION
  BEGIN ROLLBACK; RESIGNAL; END;
  START TRANSACTION;
  DELETE FROM category WHERE id = cid;
  UPDATE bookmark SET category_id = 0 WHERE category_id = cid;
  COMMIT;
END;;

CREATE PROCEDURE reorder (uid INT, new_id INT, old_id INT)
BEGIN
  DECLARE EXIT HANDLER FOR SQLEXCEPTION
  BEGIN ROLLBACK; RESIGNAL; END;
  START TRANSACTION;
  SET @new_seq := (SELECT seq FROM seq WHERE bookmark_id = new_id);
  SET @old_seq := (SELECT seq FROM seq WHERE bookmark_id = old_id);
  IF @old_seq > @new_seq
  THEN UPDATE seq SET seq = seq + 1 WHERE seq >= @new_seq AND seq < @old_seq AND user_id = uid;
  ELSE UPDATE seq SET seq = seq - 1 WHERE seq > @old_seq AND seq <= @new_seq AND user_id = uid;
  END IF;
  UPDATE seq SET seq = @new_seq WHERE bookmark_id = old_id;
  COMMIT;
END;;

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
