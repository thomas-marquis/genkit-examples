# /// script
# dependencies = [
#   "streamlit",
# ]
# ///

import streamlit as st
import json
import os

def load_data():
    file_path = 'temp.json'
    if not os.path.exists(file_path):
        st.error(f"File {file_path} not found.")
        return None
    with open(file_path, 'r', encoding='utf-8') as f:
        return json.load(f)

def main():
    st.set_page_config(page_title="Recipe Manager", page_icon="🍳")
    st.title("🍳 Recipe Manager & Grocery List")

    data = load_data()
    if data is None:
        return

    tab1, tab2 = st.tabs(["📖 Recipes", "🛒 Grocery List"])

    with tab1:
        st.header("Recipes")
        if "Recipes" in data:
            for recipe in data["Recipes"]:
                with st.expander(recipe.get("Title", "Untitled Recipe")):
                    col1, col2 = st.columns(2)
                    
                    with col1:
                        st.subheader("Ingredients")
                        ingredients = recipe.get("Ingredients", [])
                        if ingredients:
                            for ingredient in ingredients:
                                st.write(f"- {ingredient}")
                        else:
                            st.info("No ingredients listed.")

                    with col2:
                        st.subheader("Steps")
                        steps = recipe.get("Steps", [])
                        if steps:
                            for i, step in enumerate(steps, 1):
                                st.write(f"{i}. {step}")
                        else:
                            st.info("No steps listed.")
        else:
            st.warning("No recipes found in the data.")

    with tab2:
        st.header("Grocery List")
        grocery_list = data.get("GroceryList", [])
        if grocery_list:
            for item in grocery_list:
                st.write(f"- {item}")
        else:
            st.info("Grocery list is empty.")

if __name__ == "__main__":
    main()
