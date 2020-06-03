// model.go

package main

import (
    "database/sql"
)


type product struct {
    ID    int     `json:"id"`
    Name  string  `json:"name"`
    Description string `json:"description"`
    Price float64 `json:"price"`
    Changed int   `json:"changed"`
}

func (p *product) getProduct(db *sql.DB) error {
  return db.QueryRow("SELECT name, description, price, changed FROM products WHERE id=$1",
      p.ID).Scan(&p.Name, &p.Description, &p.Price, &p.Changed)
}

func (p *product) updateProduct(db *sql.DB) error {
    db.Exec("UPDATE products SET name=$2, description=$3, price=$4, changed=$5 WHERE id=$1 AND changed < $5",
    p.ID, p.Name, p.Description, p.Price, p.Changed)
    return p.getProduct(db);
}

func (p *product) deleteProduct(db *sql.DB) error {
  _, err := db.Exec("DELETE FROM products WHERE id=$1", p.ID)

  return err
}

func (p *product) createProduct(db *sql.DB) error {
  err := db.QueryRow(
      "INSERT INTO products(name, description, price, changed) VALUES($1, $2, $3, $4) RETURNING id",
      p.Name, p.Description, p.Price, p.Changed).Scan(&p.ID)

  if err != nil {
      return err
  }

  return nil
}

func getProducts(db *sql.DB, start, count int) ([]product, error) {
  rows, err := db.Query(
      "SELECT id, name, description, price, changed FROM products LIMIT $1 OFFSET $2",
      count, start)

  if err != nil {
      return nil, err
  }

  defer rows.Close()

  products := []product{}

  for rows.Next() {
      var p product
      if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Changed); err != nil {
          return nil, err
      }
      products = append(products, p)
  }

  return products, nil
}

func getChangedProducts(db *sql.DB, change int) ([]product, error) {
  rows, err := db.Query(
      "SELECT id, name, description, price, changed FROM products WHERE changed >= $1",
      change)

  if err != nil {
      return nil, err
  }

  defer rows.Close()

  products := []product{}

  for rows.Next() {
      var p product
      if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Changed); err != nil {
          return nil, err
      }
      products = append(products, p)
  }

  return products, nil
}