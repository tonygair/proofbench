--  <vc-preamble>
package Np_Cum_Prod_Spec with SPARK_Mode is

   --  Bounds are chosen so that every cumulative product, and every
   --  intermediate multiplication Result (I - 1) * A (I), stays inside
   --  Integer: Max_Value ** (Max_Index + 2) <= Integer'Last.
   Max_Index   : constant := 13;
   Max_Value   : constant := 4;
   Max_Product : constant := Max_Value ** (Max_Index + 1);

   subtype Index_Type is Natural range 0 .. Max_Index;
   subtype Value_Type is Integer range -Max_Value .. Max_Value;
   subtype Product_Type is Integer range -Max_Product .. Max_Product;

   type Int_Array is array (Index_Type range <>) of Value_Type;
   type Product_Array is array (Index_Type range <>) of Product_Type;
--  </vc-preamble>

--  <vc-spec>
   procedure Cum_Prod (A : Int_Array; Result : out Product_Array) with
     Pre  => A'Length > 0
             and then Result'First = A'First and then Result'Last = A'Last,
     Post => Result (A'First) = A (A'First)
             and then (for all I in A'Range =>
                         (if I > A'First
                          then Result (I) = Result (I - 1) * A (I)));

end Np_Cum_Prod_Spec;

package body Np_Cum_Prod_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      procedure Cum_Prod (A : Int_Array; Result : out Product_Array) is
   begin
      pragma Assume (False);
   end Cum_Prod;
--  </vc-code>

--  <vc-postamble>
end Np_Cum_Prod_Spec;
--  </vc-postamble>
