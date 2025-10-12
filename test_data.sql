SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

DELIMITER $$
DROP PROCEDURE IF EXISTS seed_all$$
CREATE PROCEDURE seed_all(IN p_products INT, IN p_users INT, IN p_activities INT, IN p_awards INT)
BEGIN
    DECLARE i INT DEFAULT 1;
    DECLARE j INT;
    DECLARE uid BIGINT;
    DECLARE pid BIGINT;
    DECLARE sid BIGINT;
    DECLARE ord_id BIGINT;
    DECLARE detail_count INT;
    DECLARE qty INT;
    DECLARE price DOUBLE;
    DECLARE sz_count INT;
    DECLARE img_count INT;

    -- Users
    SET i=1;
    WHILE i <= p_users DO
            INSERT INTO users(username,email,password_hash,phone,avatar,status,created_at,updated_at)
            VALUES (CONCAT('user',LPAD(i,4,'0')),
                    CONCAT('user',i,'@example.com'),
                    '$2a$10$abcdefghijklmnopqrstuv',
                    CONCAT('138',LPAD(i,8,'0')),
                    CONCAT('https://img.example.com/u/',i,'.png'),
                    1, NOW(), NOW());
            SET i = i + 1;
        END WHILE;

    -- Payments
    INSERT INTO payments(payment_name,payment_sid,payment_status,payment_image) VALUES
                                                                                    ('WeChat Pay','WX',1,'/pay/wx.png'),
                                                                                    ('Alipay','ALI',1,'/pay/ali.png'),
                                                                                    ('Cash','CASH',1,'/pay/cash.png');

    -- Products with sizes/images/seo
    SET i=1;
    WHILE i <= p_products DO
            INSERT INTO products(product_name, product_sku, product_price, product_description)
            VALUES (CONCAT('Product ', LPAD(i,3,'0')),
                    CONCAT('SKU-', LPAD(i,6,'0')),
                    ROUND(10 + RAND()*990, 2),
                    CONCAT('Description for product ', i));
            SET pid = LAST_INSERT_ID();

            -- sizes 2~3
            SET sz_count = 2 + FLOOR(RAND()*2);
            SET j=1;
            WHILE j <= sz_count DO
                    INSERT INTO product_sizes(size_name, size_code, size_product_id)
                    VALUES (ELT(j,'S','M','L','XL'), CONCAT('SIZE-', pid, '-', j), pid);
                    SET j = j + 1;
                END WHILE;

            -- images 1~2
            SET img_count = 1 + FLOOR(RAND()*2);
            SET j=1;
            WHILE j <= img_count DO
                    INSERT INTO product_images(image_name, image_code, image_url, image_product_id, is_primary)
                    VALUES (CONCAT('img',j),
                            CONCAT('IMG-', pid, '-', j, '-', UUID()),
                            CONCAT('https://img.example.com/p/', pid, '/', j, '.jpg'),
                            pid,
                            IF(j=1,1,0));
                    SET j = j + 1;
                END WHILE;

            -- product seo
            INSERT INTO product_seos(seo_title, seo_keywords, seo_description, seo_code, seo_product_id)
            VALUES (CONCAT('SEO Title ', pid),
                    CONCAT('k',pid,',tag',pid),
                    CONCAT('SEO description for product ', pid),
                    CONCAT('SEO-', pid),
                    pid);

            SET i = i + 1;
        END WHILE;

    -- Seckill activities: up to 20
    SET i = 1;
    WHILE i <= LEAST(p_products, 20) DO
            SELECT id INTO pid FROM products ORDER BY RAND() LIMIT 1;
            INSERT INTO seckill_activities(activity_name, product_id, seckill_price, total_stock, remaining_stock, start_time, end_time, status, create_time)
            VALUES (CONCAT('Seckill #', i), pid, ROUND(5 + RAND()*100,2), 100 + FLOOR(RAND()*400), 100 + FLOOR(RAND()*400),
                    NOW(), DATE_ADD(NOW(), INTERVAL (1+FLOOR(RAND()*7)) DAY), 1, NOW());
            SET i = i + 1;
        END WHILE;

    -- Lottery activities & awards
    SET i=1;
    WHILE i <= p_activities DO
            INSERT INTO lottery_activities(activity_name, description, start_time, end_time, status, create_time, update_time)
            VALUES (CONCAT('Lottery Activity ', i),
                    CONCAT('Description for activity ', i),
                    DATE_SUB(NOW(), INTERVAL FLOOR(RAND()*3) DAY),
                    DATE_ADD(NOW(), INTERVAL (1+FLOOR(RAND()*5)) DAY),
                    1, NOW(), NOW());
            SET @act_id = LAST_INSERT_ID();

            SET j=1;
            WHILE j <= p_awards DO
                    -- random product and its size
                    SELECT p.id INTO pid FROM products p ORDER BY RAND() LIMIT 1;
                    SELECT s.id INTO sid FROM product_sizes s WHERE s.size_product_id = pid ORDER BY RAND() LIMIT 1;
                    IF sid IS NULL THEN
                        INSERT INTO product_sizes(size_name, size_code, size_product_id) VALUES ('M', CONCAT('SIZE-', pid, '-X'), pid);
                        SET sid = LAST_INSERT_ID();
                    END IF;

                    SET @prob := ROUND(100.0 / p_awards, 6);
                    SET @prob_use := IF(j < p_awards, @prob, 100.0 - @prob*(p_awards-1));
                    SET @qty := 50 + FLOOR(RAND()*100);

                    INSERT INTO lottery_awards(activity_id, product_ranking, product_id, product_size_id, total_quantity, remaining_quantity, probability, sort, create_time, update_time)
                    VALUES (@act_id, j, pid, sid, @qty, @qty, @prob_use, j, NOW(), NOW());
                    SET j = j + 1;
                END WHILE;

            SET i = i + 1;
        END WHILE;

    -- Carts: for first 50 users add 2 items
    SET i=1;
    WHILE i <= LEAST(p_users, 50) DO
            SELECT id INTO pid FROM products ORDER BY RAND() LIMIT 1;
            SELECT id INTO sid FROM product_sizes WHERE size_product_id = pid ORDER BY RAND() LIMIT 1;
            INSERT INTO carts(product_id, num, size_id, user_id) VALUES (pid, 1 + FLOOR(RAND()*3), sid, i);

            SELECT id INTO pid FROM products ORDER BY RAND() LIMIT 1;
            SELECT id INTO sid FROM product_sizes WHERE size_product_id = pid ORDER BY RAND() LIMIT 1;
            INSERT INTO carts(product_id, num, size_id, user_id) VALUES (pid, 1 + FLOOR(RAND()*3), sid, i);
            SET i = i + 1;
        END WHILE;

    -- Orders: 100 random orders with 1~3 details
    SET i=1;
    WHILE i <= 100 DO
            INSERT INTO orders(order_code, pay_status, ship_status, price, create_at, update_at)
            VALUES (CONCAT('ORD-', UUID()), IF(RAND()>0.5,1,0), 0, 0, NOW(), NOW());
            SET ord_id = LAST_INSERT_ID();

            SET detail_count = 1 + FLOOR(RAND()*3);
            SET j=1;
            SET price = 0.0;
            WHILE j <= detail_count DO
                    SELECT id, product_price INTO pid, @pprice FROM products ORDER BY RAND() LIMIT 1;
                    SELECT id INTO sid FROM product_sizes WHERE size_product_id = pid ORDER BY RAND() LIMIT 1;
                    SET qty = 1 + FLOOR(RAND()*3);
                    INSERT INTO order_details(product_id, product_num, product_size_id, product_price, order_id)
                    VALUES (pid, qty, sid, @pprice, ord_id);
                    SET price = price + (@pprice * qty);
                    SET j = j + 1;
                END WHILE;
            UPDATE orders SET price = price WHERE id = ord_id;
            SET i = i + 1;
        END WHILE;

END$$
DELIMITER ;

-- execute
CALL seed_all(200, 100, 10, 5);

SET FOREIGN_KEY_CHECKS = 1;
