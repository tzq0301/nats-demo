CREATE TABLE `post`
(
    `id`      BIGINT AUTO_INCREMENT PRIMARY KEY,
    `title`   VARCHAR(255) NOT NULL DEFAULT '',
    `content` TEXT         NOT NULL
) ENGINE = 'InnoDB';